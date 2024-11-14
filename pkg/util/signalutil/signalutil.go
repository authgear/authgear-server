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

func Start(ctx context.Context, logger *log.Logger, daemons ...Daemon) {
	startCtx, cancel := context.WithCancel(ctx)
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

	// Cancel the context we pass to Start() first.
	// This causes the daemon that respects context to stop blocking and proceed to shutdown.
	cancel()

	stopCtx, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
	defer cancelTimeout()

	close(shutdown)
	waitGroup.Wait()
}
