package test

import (
	"database/sql/driver"
)

type MockDriver struct {
	Err  error
	Conn driver.Conn
}

func (d *MockDriver) Open(name string) (driver.Conn, error) {
	return d.Conn, d.Err
}
