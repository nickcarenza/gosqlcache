package test

import (
	"testing"
)

func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t}
}

type TestLogger struct {
	*testing.T
}

func (l *TestLogger) Print(v ...interface{}) {
	l.T.Log(v...)
}
func (l *TestLogger) Printf(format string, v ...interface{}) {
	l.T.Logf(format, v...)
}
func (l *TestLogger) Println(v ...interface{}) {
	l.T.Log(v...)
}
