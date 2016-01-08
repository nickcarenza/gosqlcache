package gosqlcache

import (
	"github.com/astaxie/beego/cache"
	"time"
)

type SecondsTimeoutCacheWrapper struct {
	cache.Cache
}

func (s *SecondsTimeoutCacheWrapper) Put(key string, val interface{}, timeout time.Duration) error {
	return s.Cache.Put(key, val, int64(timeout.Seconds()))
}
