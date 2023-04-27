package accountanonymization

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/util/backgroundjob"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func NewRunner(loggerFactory *log.Factory, runnableFactory backgroundjob.RunnableFactory) *backgroundjob.Runner {
	return backgroundjob.NewRunner(
		loggerFactory.New("account-anonymization-runner"),
		runnableFactory,
	)
}

var DependencySet = wire.NewSet(
	NewRunnableFactory,
	NewRunner,
)
