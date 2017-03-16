package gosqlcache

import (
	"net"
	"time"
)

// TODO obsolete i think

type defaultDialer struct{}

func (d defaultDialer) Dial(ntw, addr string) (net.Conn, error) {
	return net.Dial(ntw, addr)
}
func (d defaultDialer) DialTimeout(ntw, addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(ntw, addr, timeout)
}
