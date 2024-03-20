package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type Spec struct {
	Name          string
	ListenAddress string
	HTTPS         bool
	CertFilePath  string
	KeyFilePath   string
	Handler       http.Handler

	server *http.Server
}

func NewSpec(spec *Spec) *Spec {
	spec.server = &http.Server{
		Addr:              spec.ListenAddress,
		Handler:           spec.Handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return spec
}

func (spec *Spec) DisplayName() string {
	return spec.Name
}

func (spec *Spec) Start(_ context.Context, logger *log.Logger) {
	var err error
	if spec.HTTPS {
		logger.Infof("starting %v on https://%v", spec.Name, spec.ListenAddress)
		err = spec.server.ListenAndServeTLS(spec.CertFilePath, spec.KeyFilePath)
	} else {
		logger.Infof("starting %v on http://%v", spec.Name, spec.ListenAddress)
		err = spec.server.ListenAndServe()
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Fatalf("failed to start %v", spec.Name)
	}
}

func (spec *Spec) Stop(ctx context.Context, logger *log.Logger) error {
	return spec.server.Shutdown(ctx)
}
