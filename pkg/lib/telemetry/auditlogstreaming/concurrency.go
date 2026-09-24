package auditlogstreaming

import "sync"

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
func forEachBounded[T any](items []T, maxConcurrent int, fn func(item T)) {
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(item T) {
			defer wg.Done()
			defer func() { <-sem }()
			fn(item)
		}(item)
	}
	wg.Wait()
}
