package gosqlcache

import (
	"database/sql/driver"
	"io/ioutil"
	"log"
	"time"
)

import . "./interfaces"

func NewSqlCacheDriver(d driver.Driver, c SqlCacher, l Logger) *SqlCacheDriver {
	if l == nil {
		l = log.New(ioutil.Discard, "", 0)
	}
	return &SqlCacheDriver{
		Driver:    d,
		SqlCacher: c,
		Logger:    l,
		cacheConn: &cacheConn{
			cache:    c,
			cacheMap: map[string]time.Duration{},
			log:      l,
		},
	}
}

type SqlCacheDriver struct {
	driver.Driver
	SqlCacher
	Logger
	*cacheConn
}

func (d *SqlCacheDriver) Open(name string) (driver.Conn, error) {
	return Open(d, name)
}

func Open(d *SqlCacheDriver, name string) (*cacheConn, error) {
	cn, err := d.Driver.Open(name)

	if err != nil {
		return nil, err
	}

	d.cacheConn.Conn = cn
	d.cacheConn.Queryer = cn.(driver.Queryer)
	d.cacheConn.Execer = cn.(driver.Execer)

	return d.cacheConn, nil
}
