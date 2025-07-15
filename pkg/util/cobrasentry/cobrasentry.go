package cobrasentry

import (
	"context"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/util/cobraviper"
	"github.com/authgear/authgear-server/pkg/util/panicutil"
	sentryutil "github.com/authgear/authgear-server/pkg/util/sentry"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var ArgSentryDSN = &cobraviper.StringArgument{
	ArgumentName: "sentry-dsn",
	EnvName:      "SENTRY_DSN",
	Usage:        "The sentry DSN for reporting command error",
}

type BinderGetter func() *cobraviper.Binder

type InputRunEFunc func(ctx context.Context, cmd *cobra.Command, args []string) error
type OutputRunEFunc func(cmd *cobra.Command, args []string) error

var CobrasentryLogger = slogutil.NewLogger("cobra-sentry")

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
		logger := CobrasentryLogger.GetLogger(ctx).With(
			slog.String("cmd_name", cmd.Name()),
			slog.String("cmd_short", cmd.Short),
		)

		defer func() {
			e := recover()
			if e != nil {
				err = panicutil.MakeError(e)
				logger.WithError(err).Error(ctx, "panic occurred")
			}
			if sentryHub != nil {
				sentryHub.Flush(2 * time.Second)
			}
		}()

		err = do(ctx, cmd, args)
		if err != nil {
			logger.WithError(err).Error(ctx, "command exit")
		}
		return err
	}
}
