package test

import (
	"database/sql/driver"
	"time"
)

import . "../interfaces"

func SpyOnSqlCacher(c SqlCacher) *SqlCacheSpy {
	return &SqlCacheSpy{
		SqlCacher: c,
	}
}

type SqlCacheSpy struct {
	Spy
	SqlCacher
}

func (c *SqlCacheSpy) GetQueryRows(query string, args []driver.Value) driver.Rows {
	c.Call("GetQueryRows", []interface{}{query, args})
	return c.SqlCacher.GetQueryRows(query, args)
}

func (c *SqlCacheSpy) PutQueryRows(query string, args []driver.Value, val driver.Rows, timeout time.Duration) error {
	c.Call("PutQueryRows", []interface{}{query, args, val, timeout})
	return c.SqlCacher.PutQueryRows(query, args, val, timeout)
}

func (c *SqlCacheSpy) IsExistQueryRows(query string, args []driver.Value) bool {
	c.Call("IsExistQueryRows", []interface{}{query, args})
	return c.SqlCacher.IsExistQueryRows(query, args)
}

func (c *SqlCacheSpy) DeleteQueryRows(query string, args []driver.Value) error {
	c.Call("DeleteQueryRows", []interface{}{query, args})
	return c.SqlCacher.DeleteQueryRows(query, args)
}

func (c *SqlCacheSpy) ClearAll() error {
	c.Call("ClearAll", []interface{}{})
	return c.SqlCacher.ClearAll()
}
