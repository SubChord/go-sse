package sse

import (
	"log"
	"sync"
)

type Logger interface {
	Errorf(format string, v ...interface{})
}

var logger = newDummyLogger()

type dummyLoggerType struct{}

func newDummyLogger() Logger {
	return dummyLogger
}

func (d *dummyLoggerType) Errorf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

var dummyLogger = &dummyLoggerType{}

var setLoggerOnce sync.Once

func InitLogger(l Logger) {
	setLoggerOnce.Do(func() {
		logger = l
	})
}
