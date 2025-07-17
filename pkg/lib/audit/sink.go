package audit

import (
	"context"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var SinkLogger = slogutil.NewLogger("audit-sink")

type Sink struct {
	Database *auditdb.WriteHandle
	Store    *WriteStore
}

func (s *Sink) ReceiveBlockingEvent(ctx context.Context, e *event.Event) (err error) {
	// We do not log blocking event.
	return
}

func (s *Sink) ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) (err error) {
	logger := SinkLogger.GetLogger(ctx)
	// Skip events that are not for audit.
	payload := e.Payload.(event.NonBlockingPayload)
	if !payload.ForAudit() {
		return
	}

	if s.Database == nil {
		logger.Debug(ctx, "skip persisting event", slog.Any("event", e))
		return
	}

	logger.Debug(ctx, "persisting event", slog.Any("event", e))

	err = s.Database.WithTx(ctx, func(ctx context.Context) error {
		logEntry, err := NewLog(e)
		if err != nil {
			return err
		}
		return s.Store.PersistLog(ctx, logEntry)
	})
	if err != nil {
		return
	}

	return
}
