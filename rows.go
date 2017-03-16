package gosqlcache

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"io"
)

// Implements driver.Rows
type cachedRows struct {
	*stmt
	cols    []string
	pointer int
	data    [][]driver.Value
	closed  bool
}

func newCachedRows(dr driver.Rows) (r *cachedRows, err error) {
	r = &cachedRows{}
	r.cols = dr.Columns()
	for {
		var cols = make([]driver.Value, len(r.cols))
		err = dr.Next(cols)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		r.data = append(r.data, cols)
	}
}

func (r *cachedRows) Columns() []string {
	return r.cols
}

func (r *cachedRows) Close() error {
	r.data = nil
	r.closed = true
	return nil
}

func (r *cachedRows) Next(dest []driver.Value) error {
	if r.closed {
		return driver.ErrBadConn
	}
	if len(r.data) <= r.pointer {
		return io.EOF
	}
	for i, v := range r.data[r.pointer] {
		dest[i] = v
	}
	r.pointer += 1
	return nil
}

func (r *cachedRows) GobEncode() ([]byte, error) {
	r.stmt.cacheConn.log.Println("GobEncode")
	var buf bytes.Buffer
	var enc = gob.NewEncoder(&buf)
	// r.stmt.cacheConn.log.Printf("Encoding cols %#v", r.cols)
	// r.stmt.cacheConn.log.Printf("Encoding data %#v", r.data)
	err := enc.Encode(map[string]interface{}{
		"columns": r.cols,
		"data":    r.data,
	})
	if err != nil {
		r.stmt.cacheConn.log.Println("Unable to encode cached rows: ", err)
	} else {
		// r.stmt.cacheConn.log.Println("Encoded to: ", buf.String())
		// r.stmt.cacheConn.log.Printf("Encoded to: %#v\n", buf.Bytes())
	}
	return buf.Bytes(), err
}

func (r *cachedRows) GobDecode(b []byte) (err error) {
	r.stmt.cacheConn.log.Println("GobDecode")
	// r.stmt.cacheConn.log.Printf("Decoding from: %#v\n", b)
	var buf = bytes.NewBuffer(b)
	var dec = gob.NewDecoder(buf)
	var m = map[string]interface{}{}
	err = dec.Decode(&m)
	if err != nil {
		r.stmt.cacheConn.log.Println("Unable to decode", err)
		return
	}
	// r.stmt.cacheConn.log.Printf("Decoded to: %#v", m)
	var ok bool
	r.cols, ok = m["columns"].([]string)
	if !ok {
		r.cols = []string{}
	}
	r.data, ok = m["data"].([][]driver.Value)
	if !ok {
		r.data = [][]driver.Value{}
	}
	return nil
}
