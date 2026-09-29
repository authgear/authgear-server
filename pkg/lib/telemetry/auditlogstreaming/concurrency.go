package auditlogstreaming

import (
	"context"
	"sync"

	"github.com/authgear/authgear-server/pkg/util/panicutil"
)

// forEachBounded runs fn once per item in items, with at most maxConcurrent
// calls running at a time, and blocks until every call has returned.
//
// Used to bound how much of a tick's work (every app's drain, every
// stream's delivery) runs concurrently: without a bound, either running
// everything sequentially makes one slow/unreachable collector serially
// delay every other app or stream, or running everything at once opens
// unbounded outbound connections. A bounded pool keeps a tick's wall-clock
// time close to (item count / maxConcurrent) * slowest item, instead of
// (item count) * slowest item, which matters because the drain lock this
// package holds for the whole tick (see runnable.go's WithMutexExpiry
// call) can otherwise expire mid-tick and let a second replica start
// draining concurrently.
//
// A panic in fn is recovered and logged: fn runs on its own goroutine, so
// backgroundjob.Runner's recover cannot catch it and it would otherwise
// crash the whole process.
func forEachBounded[T any](ctx context.Context, items []T, maxConcurrent int, fn func(item T)) {
	logger := Logger.GetLogger(ctx)
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(item T) {
			defer wg.Done()
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil {
					logger.WithError(panicutil.MakeError(r)).Error(ctx, "panic occurred in audit log streaming")
				}
			}()
			fn(item)
		}(item)
	}
	wg.Wait()
}
