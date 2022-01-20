package accountdeletion

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func NewRunner(loggerFactory *log.Factory, runnable backgroundjob.Runnable) *backgroundjob.Runner {
	return backgroundjob.NewRunner(
		loggerFactory.New("account-deletion-runner"),
		runnable,
	)
}

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Runnable), "*"),
	NewRunner,
	NewRunnableLogger,
	wire.Bind(new(backgroundjob.Runnable), new(*Runnable)),
)
