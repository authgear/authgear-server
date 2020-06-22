package interaction

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("interaction")}
}

var DependencySet = wire.NewSet(
	NewLogger,
	wire.Struct(new(Provider), "*"),
)
