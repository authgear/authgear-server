package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("server")

type Spec struct {
	Name          string
	ListenAddress string
	HTTPS         bool
	CertFilePath  string
	KeyFilePath   string
	Handler       http.Handler

	server *http.Server
}

func NewSpec(ctx context.Context, spec *Spec) *Spec {
	spec.server = &http.Server{
		Addr:              spec.ListenAddress,
		Handler:           spec.Handler,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	return spec
}

func (spec *Spec) DisplayName() string {
	return spec.Name
}

func (spec *Spec) Start(ctx context.Context) {
	logger := logger.GetLogger(ctx)
	var err error
	if spec.HTTPS {
		logger.Info(ctx, "starting on https", slog.String("name", spec.Name), slog.String("listen_address", spec.ListenAddress))
		err = spec.server.ListenAndServeTLS(spec.CertFilePath, spec.KeyFilePath)
	} else {
		logger.Info(ctx, "starting on http", slog.String("name", spec.Name), slog.String("listen_address", spec.ListenAddress))
		err = spec.server.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Error(ctx, "failed to start", slog.String("name", spec.Name))
		panic(err)
	}
}

func (spec *Spec) Stop(ctx context.Context) error {
	return spec.server.Shutdown(ctx)
}
