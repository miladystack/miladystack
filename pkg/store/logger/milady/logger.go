package milady

import (
	"github.com/miladystack/miladystack/pkg/log"
)

type miladyLogger struct{}

func NewLogger() *miladyLogger {
	return &miladyLogger{}
}

func (l *miladyLogger) Error(err error, msg string, kvs ...any) {
	log.Errorw(err, msg, kvs...)
}
