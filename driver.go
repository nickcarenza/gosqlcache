package gosqlcache

import (
	"database/sql/driver"
	"io/ioutil"
	"log"
	"time"
)

func NewSqlCacheDriver(d driver.Driver, c SqlCacher, l Logger) SqlCacheDriver {
	if l == nil {
		l = log.New(ioutil.Discard, "", 0)
	}
	return SqlCacheDriver{d, c, l}
}

type SqlCacheDriver struct {
	driver.Driver
	SqlCacher
	Logger
}

func (d *SqlCacheDriver) Open(name string) (driver.Conn, error) {
	return Open(d, name)
}

func Open(d *SqlCacheDriver, name string) (*cacheConn, error) {
	cn, err := d.Driver.Open(name)

	if err != nil {
		return nil, err
	}

	return &cacheConn{
		Conn:     cn,
		Queryer:  cn.(driver.Queryer),
		Execer:   cn.(driver.Execer),
		cache:    d.SqlCacher,
		cacheMap: map[string]time.Duration{},
		log:      d.Logger,
	}, nil
}
