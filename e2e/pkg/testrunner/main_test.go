package testrunner

import (
	"testing"
)

func TestAuthflow(t *testing.T) {
	runner := NewTestRunner(t, "../../tests/")
	runner.Run()
}
