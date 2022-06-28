package cobrasentry

import (
	"github.com/getsentry/sentry-go"

	"github.com/authgear/authgear-server/pkg/util/log"
	sentryutil "github.com/authgear/authgear-server/pkg/util/sentry"
)

func NewLoggerFactory(sentryHub *sentry.Hub) *log.Factory {
	logLevel := log.LevelInfo
	if sentryHub == nil {
		return log.NewFactory(logLevel)
	}

	hub := sentryutil.NewLogHookFromHub(sentryHub)
	return log.NewFactory(
		logLevel,
		hub,
	)
}
