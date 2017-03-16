package test

import (
	"database/sql/driver"
)

type MockRows struct {
	Cols     []string
	Err      error
	NextVals []driver.Value
}

func (r *MockRows) Columns() []string {
	return r.Cols
}

func (r *MockRows) Close() error {
	return r.Err
}

func (r *MockRows) Next(dest []driver.Value) error {
	for i := range dest {
		dest[i] = r.NextVals[i]
	}
	return r.Err
}
