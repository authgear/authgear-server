package audit

import (
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

// MockTrail is a mock audit trail with testing logger
type MockTrail struct {
	logger *logrus.Logger
	Hook   *test.Hook
}

func (t MockTrail) Log(entry Entry) {
	t.logger.WithFields(entry.toLogrusFields()).Info("audit_trail")
}

func (t MockTrail) WithRequest(req *http.Request) Trail {
	fields := logrus.Fields{}
	fields["remote_addr"] = req.RemoteAddr
	fields["x_forwarded_for"] = req.Header.Get("x-forwarded-for")
	fields["x_real_ip"] = req.Header.Get("x-real-ip")
	fields["forwarded"] = req.Header.Get("forwarded")

	return &LoggerTrail{
		logger: t.logger.WithFields(fields),
	}
}

// NewMockTrail create mock audit trail with testing logger
func NewMockTrail(t *testing.T) *MockTrail {
	logger, hook := test.NewNullLogger()
	return &MockTrail{
		logger: logger,
		Hook:   hook,
	}
}
