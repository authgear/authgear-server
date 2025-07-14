package backgroundjob

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

var logger = slogutil.NewLogger("background")

// Main takes a list of runners and start them.
// Upon receiving SIGINT or SIGTERM, stop them gracefully.
func Main(ctx context.Context, runners []*Runner) {
	logger := logger.GetLogger(ctx)
	var waitGroup sync.WaitGroup
	shutdown := make(chan struct{})

	for _, runner := range runners {
		// Capture
		runner := runner

		go runner.Start()
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-shutdown
			runner.Stop(ctx)
		}()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Info(ctx, "received signal, shutting down...", slog.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	close(shutdown)
	waitGroup.Wait()
}
