package audit

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("audit-sink")} }

type Sink struct {
	Logger   Logger
	Database *auditdb.WriteHandle
	Store    *WriteStore
}

func (s *Sink) ReceiveBlockingEvent(e *event.Event) (err error) {
	// We do not log blocking event.
	return
}

func (s *Sink) ReceiveNonBlockingEvent(e *event.Event) (err error) {
	// Skip events that are not for audit.
	payload := e.Payload.(event.NonBlockingPayload)
	if !payload.ForAudit() {
		return
	}

	if s.Database == nil {
		s.Logger.WithFields(map[string]interface{}{
			"event": e,
		}).Debug("skip persisting event")
		return
	}

	s.Logger.WithFields(map[string]interface{}{
		"event": e,
	}).Debug("persisting event")

	err = s.Database.WithTx(func() error {
		logEntry, err := NewLog(e)
		if err != nil {
			return err
		}
		return s.Store.PersistLog(logEntry)
	})
	if err != nil {
		return
	}

	return
}
