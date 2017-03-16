package gosqlcache

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

import . "./interfaces"

func NewSqlCacher(c Cacher, l Logger) *SqlCache {
	if l == nil {
		l = log.New(ioutil.Discard, "", 0)
	}
	return &SqlCache{c, l}
}

type SqlCache struct {
	Cacher
	Logger
}

func (c *SqlCache) GetQueryRows(query string, args []driver.Value) driver.Rows {
	key := getCacheKey(query, args)
	cachedValue := c.Cacher.Get(key)
	if cachedValue == nil {
		return nil
	}
	var reader = bytes.NewReader(cachedValue.([]byte))
	decoder := gob.NewDecoder(reader)
	r := cachedRows{}
	err := decoder.Decode(&r)
	if err != nil {
		c.Println("Unable to decode cached query rows: ", err)
	}
	return driver.Rows(&r)
}

func (c *SqlCache) PutQueryRows(query string, args []driver.Value, val driver.Rows, timeout time.Duration) (err error) {
	c.Println("Caching query rows")
	key := getCacheKey(query, args)
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)
	err = enc.Encode(val)
	if err != nil {
		return
	}
	// c.Printf("Putting query rows: %#v\n", buf.Bytes())
	// c.Printf("Putting query rows: %#v\n", buf.String())
	return c.Cacher.Put(key, buf.Bytes(), timeout)
}

func (c *SqlCache) IsExistQueryRows(query string, args []driver.Value) bool {
	key := getCacheKey(query, args)
	return c.Cacher.IsExist(key)
}

func (c *SqlCache) DeleteQueryRows(query string, args []driver.Value) (err error) {
	key := getCacheKey(query, args)
	return c.Cacher.Delete(key)
}

func getCacheKey(query string, args []driver.Value) string {
	keyparts := []string{query}
	for _, v := range args {
		keyparts = append(keyparts, fmt.Sprintf("%[1]T:%[1]v", v))
	}
	return strings.Join(keyparts, ` `)
}
