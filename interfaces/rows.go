package interfaces

import (
	"database/sql/driver"
	"encoding/gob"
)

type Rows interface {
	gob.GobEncoder
	gob.GobDecoder
	driver.Rows
}
