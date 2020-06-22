package redis

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("redis-session-store")}
}

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Store), "*"),
)
