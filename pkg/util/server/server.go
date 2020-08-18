package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
}

func Start(logger *log.Logger, specs []Spec) {
	var ctx context.Context
	waitGroup := new(sync.WaitGroup)
	shutdown := make(chan struct{})

	for _, spec := range specs {
		// Capture spec
		spec := spec

		httpServer := &http.Server{
			Addr:    spec.ListenAddress,
			Handler: spec.Handler,
		}

		go func() {
			logger.Infof("starting %v on %v", spec.Name, spec.ListenAddress)
			var err error
			if spec.HTTPS {
				err = httpServer.ListenAndServeTLS(spec.CertFilePath, spec.KeyFilePath)
			} else {
				err = httpServer.ListenAndServe()
			}

			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.WithError(err).Fatalf("failed to start %v", spec.Name)
			}
		}()

		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			select {
			case <-shutdown:
				break
			}

			logger.Infof("stopping %v...", spec.Name)

			err := httpServer.Shutdown(ctx)
			if err != nil {
				logger.WithError(err).Errorf("failed to stop gracefully %v", spec.Name)
			}
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sig:
		logger.Infof("received signal %s, shutting down...", sig.String())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	close(shutdown)
	waitGroup.Wait()
}
