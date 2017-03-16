package test

import (
	"database/sql/driver"
)

type MockConn struct {
	Err    error
	Stmt   driver.Stmt
	Tx     driver.Tx
	Rows   driver.Rows
	Result driver.Result
	Spy
}

func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	c.Call("Prepare", []interface{}{query})
	return c.Stmt, c.Err
}

func (c *MockConn) Close() error {
	c.Call("Close", []interface{}{})
	return c.Err
}

func (c *MockConn) Begin() (driver.Tx, error) {
	c.Call("Begin", []interface{}{})
	return c.Tx, c.Err
}

func (c *MockConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	c.Call("Query", []interface{}{query, args})
	return c.Rows, c.Err
}

func (c *MockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	c.Call("Exec", []interface{}{query, args})
	return c.Result, c.Err
}
