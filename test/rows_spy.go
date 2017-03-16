package test

import (
	"database/sql/driver"
)

import . "../interfaces"

func SpyOnRows(r Rows) *RowsSpy {
	return &RowsSpy{
		Rows: r,
	}
}

type RowsSpy struct {
	Spy
	Rows
}

func (r *RowsSpy) Columns() []string {
	r.Call("Columns", []interface{}{})
	return r.Rows.Columns()
}

func (r *RowsSpy) Close() error {
	r.Call("Close", []interface{}{})
	return r.Rows.Close()
}

func (r *RowsSpy) Next(dest []driver.Value) error {
	r.Call("Next", []interface{}{dest})
	return r.Rows.Next(dest)
}
