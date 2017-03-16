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
}

func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	return c.Stmt, c.Err
}

func (c *MockConn) Close() error {
	return c.Err
}

func (c *MockConn) Begin() (driver.Tx, error) {
	return c.Tx, c.Err
}

func (c *MockConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	return c.Rows, c.Err
}

func (c *MockConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	return c.Result, c.Err
}
