package gosqlcache

import (
	"database/sql/driver"
	"time"
)

// Implements driver.Conn
type cacheConn struct {
	driver.Conn
	driver.Queryer
	driver.Execer
	cache          SqlCacher
	pendingQueries []stmt
	cacheMap       map[string]time.Duration
	log            Logger
}

func (c *cacheConn) Prepare(query string) (driver.Stmt, error) {
	s, err := c.Conn.Prepare(query)
	return &stmt{query, c, s}, err
}

// make stmt then execute statement
func (c *cacheConn) Query(query string, args []driver.Value) (r driver.Rows, err error) {
	s, err := c.Prepare(query)
	if err != nil {
		return
	}
	return s.Query(args)
}

func (c *cacheConn) RegisterQuery(query string, d time.Duration) {
	c.cacheMap[query] = d
}

func (c *cacheConn) SetLogger(l Logger) {
	c.log = l
}

func (c *cacheConn) SetCacher(cache Cacher) {
	c.cache = &SqlCache{cache, c.log}
}

func (c *cacheConn) SetSqlCacher(cache SqlCacher) {
	c.cache = cache
}
