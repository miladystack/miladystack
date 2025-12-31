package milady

import (
	"context"

	"github.com/miladystack/miladystack/pkg/log"
)

type miladyLogger struct{}

func NewLogger() *miladyLogger {
	return &miladyLogger{}
}

func (l *miladyLogger) Error(ctx context.Context, err error, msg string, kvs ...any) {
	log.Errorw(err, msg, kvs...)
}
