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
	Database *auditdb.Handle
	Store    *Store
}

func (s *Sink) ReceiveBlockingEvent(e *event.Event) (err error) {
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
		return s.Store.PersistEvent(e)
	})
	if err != nil {
		return
	}

	return
}

func (s *Sink) ReceiveNonBlockingEvent(e *event.Event) (err error) {
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
		return s.Store.PersistEvent(e)
	})
	if err != nil {
		return
	}

	return
}
