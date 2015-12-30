package gosqlcache

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// TODO synchronize multiple queries to the same result set and only run it once
// TODO allow multiple connection to have different caches
// TODO add stats

var ERR_CLOSED = errors.New("Closed")

var _cache sqlcacher

// Queries to cache
var _cacheMap = map[string]time.Duration{}

var _log logger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

type logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

func SetLogger(l logger) {
	_log = l
}

func RegisterQuery(query string, d time.Duration) {
	_cacheMap[query] = d
}

type drv struct{}

func (d *drv) Open(name string) (driver.Conn, error) {
	return Open(name)
}

func init() {
	_log.Println("init")
	sql.Register("postgres-cached", &drv{})
	gob.Register([][]driver.Value{})
}

func Open(name string) (_ driver.Conn, err error) {
	cn, err := pq.DialOpen(defaultDialer{}, name)
	return &conn{
		Conn:    cn,
		Queryer: cn.(driver.Queryer),
		Execer:  cn.(driver.Execer),
		cache:   _cache,
	}, err
}

type defaultDialer struct{}

func (d defaultDialer) Dial(ntw, addr string) (net.Conn, error) {
	return net.Dial(ntw, addr)
}
func (d defaultDialer) DialTimeout(ntw, addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(ntw, addr, timeout)
}

// Implements driver.Conn
type conn struct {
	driver.Conn
	driver.Queryer
	driver.Execer
	cache          sqlcacher
	pendingQueries []stmt
}

// Implements driver.Stmt
type stmt struct {
	query string
	*conn
	driver.Stmt
}

func (s *stmt) Query(args []driver.Value) (r driver.Rows, err error) {

	if s.conn.cache != nil {
		_log.Println("Checking cache")
		r = s.conn.cache.GetQueryRows(s.query, args)
		if r != nil {
			_log.Println("Cache hit")
			return r, nil
		}
		_log.Println("Cache miss")
	}

	r, err = s.conn.Queryer.Query(s.query, args)
	if err != nil {
		return nil, err
	}

	if d, ok := _cacheMap[s.query]; ok {
		// _err: type pq.rows has no exported fields
		cr, _err := newCachedRows(r)
		if _err != nil {
			_log.Println("Unable to cache query rows", _err)
		} else {
			cr.pointer = 0
			r = cr
			_err := s.conn.cache.PutQueryRows(s.query, args, cr, d)
			if _err != nil {
				_log.Println("Unable to cache query rows", _err)
			}
		}
	}

	return
}

// Implements driver.Rows
type cachedRows struct {
	*stmt
	cols    []string
	pointer int
	data    [][]driver.Value
	closed  bool
}

func (r *cachedRows) Columns() []string {
	return r.cols
}

func (r *cachedRows) Close() error {
	r.data = nil
	r.closed = true
	return nil
}

func (r *cachedRows) Next(dest []driver.Value) error {
	if r.closed {
		return ERR_CLOSED
	}
	if len(r.data) <= r.pointer {
		return io.EOF
	}
	for i, v := range r.data[r.pointer] {
		dest[i] = v
	}
	r.pointer += 1
	return nil
}

func (r *cachedRows) GobEncode() ([]byte, error) {
	_log.Println("GobEncode")
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)
	// _log.Printf("Encoding cols %#v", r.cols)
	// _log.Printf("Encoding data %#v", r.data)
	err := enc.Encode(map[string]interface{}{
		"columns": r.cols,
		"data":    r.data,
	})
	if err != nil {
		_log.Println("Unable to encode cached rows: ", err)
	} else {
		// _log.Println("Encoded to: ", buf.String())
		// _log.Printf("Encoded to: %#v\n", buf.Bytes())
	}
	return buf.Bytes(), err
}

func (r *cachedRows) GobDecode(b []byte) (err error) {
	_log.Println("GobDecode")
	// _log.Printf("Decoding from: %#v\n", b)
	var buf = bytes.NewBuffer(b)
	var dec = gob.NewDecoder(buf)
	var m = map[string]interface{}{}
	err = dec.Decode(&m)
	if err != nil {
		_log.Println("Unable to decode", err)
		return
	}
	// _log.Printf("Decoded to: %#v", m)
	var ok bool
	r.cols, ok = m["columns"].([]string)
	if !ok {
		r.cols = []string{}
	}
	r.data, ok = m["data"].([][]driver.Value)
	if !ok {
		r.data = [][]driver.Value{}
	}
	return nil
}

func newCachedRows(dr driver.Rows) (r *cachedRows, err error) {
	r = &cachedRows{}
	r.cols = dr.Columns()
	for {
		var cols = make([]driver.Value, len(r.cols))
		err = dr.Next(cols)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		r.data = append(r.data, cols)
	}
}

type cacher interface {
	Get(key string) interface{}
	Put(key string, val interface{}, timeout time.Duration) error
	IsExist(key string) bool
	Delete(key string) error
	ClearAll() error
}

type sqlcacher interface {
	GetQueryRows(query string, args []driver.Value) driver.Rows
	PutQueryRows(query string, args []driver.Value, val driver.Rows, timeout time.Duration) error
	IsExistQueryRows(query string, args []driver.Value) bool
	DeleteQueryRows(query string, args []driver.Value) error
	ClearAll() error
}

type sqlcache struct {
	cacher
}

func getCacheKey(query string, args []driver.Value) string {
	keyparts := []string{query}
	for _, v := range args {
		keyparts = append(keyparts, fmt.Sprintf("%[1]T:%[1]v", v))
	}
	return strings.Join(keyparts, ` `)
}

func (c *sqlcache) GetQueryRows(query string, args []driver.Value) driver.Rows {
	key := getCacheKey(query, args)
	cachedValue := c.cacher.Get(key)
	if cachedValue == nil {
		return nil
	}
	var reader = bytes.NewReader(cachedValue.([]byte))
	decoder := gob.NewDecoder(reader)
	r := cachedRows{}
	err := decoder.Decode(&r)
	if err != nil {
		_log.Println("Unable to decode cached query rows: ", err)
	}
	return driver.Rows(&r)
}

func (c *sqlcache) PutQueryRows(query string, args []driver.Value, val driver.Rows, timeout time.Duration) (err error) {
	_log.Println("Caching query rows")
	key := getCacheKey(query, args)
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)
	err = enc.Encode(val)
	if err != nil {
		return
	}
	// _log.Printf("Putting query rows: %#v\n", buf.Bytes())
	// _log.Printf("Putting query rows: %#v\n", buf.String())
	return c.cacher.Put(key, buf.Bytes(), timeout)
}

func (c *sqlcache) IsExistQueryRows(query string, args []driver.Value) bool {
	key := getCacheKey(query, args)
	return c.cacher.IsExist(key)
}

func (c *sqlcache) DeleteQueryRows(query string, args []driver.Value) (err error) {
	key := getCacheKey(query, args)
	return c.cacher.Delete(key)
}

func SetCacher(cache cacher) {
	_cache = &sqlcache{cache}
}

func SetSqlCacher(cache sqlcacher) {
	_cache = cache
}

func (cn *conn) Prepare(query string) (driver.Stmt, error) {
	s, err := cn.Conn.Prepare(query)
	return &stmt{query, cn, s}, err
}

// make stmt then execute statement
func (cn *conn) Query(query string, args []driver.Value) (r driver.Rows, err error) {
	s, err := cn.Prepare(query)
	if err != nil {
		return
	}
	return s.Query(args)
}
