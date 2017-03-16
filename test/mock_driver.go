package test

import (
	"database/sql/driver"
)

type MockDriver struct {
	Err  error
	Conn driver.Conn
	Spy
}

func (d *MockDriver) Open(name string) (driver.Conn, error) {
	d.Call("Open", []interface{}{name})
	return d.Conn, d.Err
}
