package signalutil

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/authgear/authgear-server/pkg/util/log"
)

// Daemon is something that runs indefinitely.
type Daemon interface {
	DisplayName() string
	// Start blocks.
	Start(ctx context.Context, logger *log.Logger)
	// Stop stops.
	Stop(ctx context.Context, logger *log.Logger) error
}

func Start(logger *log.Logger, daemons ...Daemon) {
	startCtx := context.Background()
	var stopCtx context.Context
	waitGroup := new(sync.WaitGroup)
	shutdown := make(chan struct{})

	for _, daemon := range daemons {
		daemon := daemon

		go daemon.Start(startCtx, logger)
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			<-shutdown

			logger.Infof("stopping %v...", daemon.DisplayName())
			err := daemon.Stop(stopCtx, logger)
			if err != nil {
				logger.WithError(err).Errorf("failed to stop gracefully %v", daemon.DisplayName())
			}
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Infof("received signal %s, shutting down...", sig.String())

	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	close(shutdown)
	waitGroup.Wait()
}
