package backgroundjob

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/authgear/authgear-server/pkg/util/log"
)

// Main takes a list of runners and start them.
// Upon receiving SIGINT or SIGTERM, stop them gracefully.
func Main(ctx context.Context, logger *log.Logger, runners []*Runner) {
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
	logger.Infof("received signal %s, shutting down...", sig.String())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	close(shutdown)
	waitGroup.Wait()
}
