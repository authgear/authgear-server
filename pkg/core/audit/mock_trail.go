package audit

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"testing"
)

// MockTrail is a mock audit trail with testing logger
type MockTrail struct {
	logger *logrus.Logger
	Hook   *test.Hook
}

func (t MockTrail) Log(entry Entry) {
	t.logger.WithFields(entry.toLogrusFields()).Info("audit_trail")
}

// NewMockTrail create mock audit trail with testing logger
func NewMockTrail(t *testing.T) *MockTrail {
	logger, hook := test.NewNullLogger()
	return &MockTrail{
		logger: logger,
		Hook:   hook,
	}
}
