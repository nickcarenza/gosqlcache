package gosqlcache

import (
	"database/sql/driver"
)

// Implements driver.Stmt
type stmt struct {
	query string
	*cacheConn
	driver.Stmt
}

func (s *stmt) Query(args []driver.Value) (r driver.Rows, err error) {

	if s.cacheConn.cache != nil {
		s.cacheConn.log.Println("Checking cache")
		r = s.cacheConn.cache.GetQueryRows(s.query, args)
		if r != nil {
			s.cacheConn.log.Println("Cache hit")
			return r, nil
		}
		s.cacheConn.log.Println("Cache miss")
	}

	r, err = s.cacheConn.Queryer.Query(s.query, args)
	if err != nil {
		return nil, err
	}

	if d, ok := s.cacheConn.cacheMap[s.query]; ok {
		// _err: type pq.rows has no exported fields
		cr, _err := newCachedRows(r)
		if _err != nil {
			s.cacheConn.log.Println("Unable to cache query rows", _err)
		} else {
			cr.pointer = 0
			r = cr
			_err := s.cacheConn.cache.PutQueryRows(s.query, args, cr, d)
			if _err != nil {
				s.cacheConn.log.Println("Unable to cache query rows", _err)
			}
		}
	}

	return
}
