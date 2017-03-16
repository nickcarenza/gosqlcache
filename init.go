package gosqlcache

import (
	"database/sql/driver"
	"encoding/gob"
)

func init() {
	gob.Register([][]driver.Value{})
}
