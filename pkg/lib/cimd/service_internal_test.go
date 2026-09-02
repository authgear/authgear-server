package cimd

import "time"

// ShrinkDocumentWaitForTest shrinks the package-level
// documentWaitPollInterval/documentWaitMaxDuration for the duration of a
// single test, so a test exercising waitForResolvedClient's timeout path
// doesn't actually block for several real seconds. Exported (but test-only,
// guarded by the _test.go suffix so it never reaches production) because
// service_test.go is package cimd_test, testing Service through its public
// API, and so cannot reach the unexported vars directly -- mirrors
// shrinkLogoWait's reasoning in logo_test.go, which can shrink its own
// package-level vars directly since logo_test.go is package cimd.
func ShrinkDocumentWaitForTest() (restore func()) {
	origInterval, origMax := documentWaitPollInterval, documentWaitMaxDuration
	documentWaitPollInterval = 2 * time.Millisecond
	documentWaitMaxDuration = 30 * time.Millisecond
	return func() {
		documentWaitPollInterval, documentWaitMaxDuration = origInterval, origMax
	}
}
