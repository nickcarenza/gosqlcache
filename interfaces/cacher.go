package interfaces

import (
	"time"
)

type Cacher interface {
	Get(key string) interface{}
	Put(key string, val interface{}, timeout time.Duration) error
	IsExist(key string) bool
	Delete(key string) error
	ClearAll() error
}
