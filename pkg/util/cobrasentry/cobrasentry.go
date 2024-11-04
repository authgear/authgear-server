package cobrasentry

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/util/cobraviper"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	sentryutil "github.com/authgear/authgear-server/pkg/util/sentry"
)

var ArgSentryDSN = &cobraviper.StringArgument{
	ArgumentName: "sentry-dsn",
	EnvName:      "SENTRY_DSN",
	Usage:        "The sentry DSN for reporting command error",
}

type BinderGetter func() *cobraviper.Binder

type InputRunEFunc func(ctx context.Context, cmd *cobra.Command, args []string) error
type OutputRunEFunc func(cmd *cobra.Command, args []string) error

var RunEWrap = func(binderGetter BinderGetter, do InputRunEFunc) OutputRunEFunc {
	return func(cmd *cobra.Command, args []string) (err error) {
		ctx := cmd.Context()

		binder := binderGetter()
		sentryDSN := binder.GetString(cmd, ArgSentryDSN)

		var sentryHub *sentry.Hub
		if sentryDSN != "" {
			sentryHub, err = sentryutil.NewHub(sentryDSN)
			if err != nil {
				return err
			}
		}

		ctx = WithHub(ctx, sentryHub)
		loggerFactory := NewLoggerFactory(sentryHub)
		logger := loggerFactory.New("cobra-sentry").
			WithField("cmd_name", cmd.Name()).
			WithField("cmd_short", cmd.Short)

		defer func() {
			e := recover()
			if e != nil {
				err = panicutil.MakeError(e)
				logger.WithError(err).
					WithField("stack", errorutil.Callers(10000)).
					Error("panic occurred")
			}
			if sentryHub != nil {
				sentryHub.Flush(2 * time.Second)
			}
		}()

		err = do(ctx, cmd, args)
		if err != nil {
			logger.WithError(err).Error("command exit")
		}
		return err
	}
}
