package gosqlcache

import (
	"database/sql/driver"
	"encoding/gob"
)

func init() {
	// _log.Println("init")

	// TODO consider making the drv public and pushing this call to the clients
	// sql.Register(DriverName, &drv{})

	gob.Register([][]driver.Value{})
}
