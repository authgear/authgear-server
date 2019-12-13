package handler

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/logging"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func reverseProxyErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	// Create a logger for the app and use it to log the error.
	ctx := model.GatewayContextFromContext(r.Context())
	tConfig := ctx.App.Config
	logHook := logging.NewDefaultLogHook(tConfig.DefaultSensitiveLoggerValues())
	// The sentry hook is not added here because the error we are logging is from upstream.
	loggerFactory := logging.NewFactoryFromRequest(r, logHook)
	logger := loggerFactory.NewLogger("deployment_route")
	logger.WithError(err).
		WithField("request_uri", r.RequestURI).
		WithField("method", r.Method).
		WithField("proto", r.Proto).
		Error("error in upstream")
	// Return 502 is the default behavior.
	w.WriteHeader(http.StatusBadGateway)
}
