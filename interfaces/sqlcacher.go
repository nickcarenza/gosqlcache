package interfaces

import (
	"database/sql/driver"
	"time"
)

type SqlCacher interface {
	GetQueryRows(query string, args []driver.Value) driver.Rows
	PutQueryRows(query string, args []driver.Value, val driver.Rows, timeout time.Duration) error
	IsExistQueryRows(query string, args []driver.Value) bool
	DeleteQueryRows(query string, args []driver.Value) error
	ClearAll() error
}
