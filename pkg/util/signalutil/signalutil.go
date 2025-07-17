package signalutil

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var logger = slogutil.NewLogger("signalutil")

// Daemon is something that runs indefinitely.
type Daemon interface {
	DisplayName() string
	// Start blocks.
	Start(ctx context.Context)
	// Stop stops.
	Stop(ctx context.Context) error
}

func Start(ctx context.Context, daemons ...Daemon) {
	logger := logger.GetLogger(ctx)

	startCtx, cancel := context.WithCancel(ctx)
	var stopCtx context.Context
	waitGroup := new(sync.WaitGroup)
	shutdown := make(chan struct{})

	for _, daemon := range daemons {
		daemon := daemon

		go daemon.Start(startCtx)
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			<-shutdown

			logger.Info(ctx, "stopping ...", slog.String("display_name", daemon.DisplayName()))
			err := daemon.Stop(stopCtx)
			if err != nil {
				logger.WithError(err).Error(ctx, "failed to stop gracefully", slog.String("display_name", daemon.DisplayName()))
			}
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info(ctx, "received signal, shutting down...", slog.String("signal", sig.String()))

	// Cancel the context we pass to Start() first.
	// This causes the daemon that respects context to stop blocking and proceed to shutdown.
	cancel()

	stopCtx, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
	defer cancelTimeout()

	close(shutdown)
	waitGroup.Wait()
}
