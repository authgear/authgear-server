package reindex

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type SinkLogger struct{ *log.Logger }

func NewSinkLogger(lf *log.Factory) SinkLogger { return SinkLogger{lf.New("search-reindex-sink")} }

type Sink struct {
	Logger    SinkLogger
	Reindexer *Reindexer
	Database  *appdb.Handle
}

func (s *Sink) ReceiveBlockingEvent(ctx context.Context, e *event.Event) error {
	return nil
}

func (s *Sink) ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) error {
	payload := e.Payload.(event.NonBlockingPayload)
	reindexRequiredUserIDs := []string{}
	reindexRequiredUserIDs = append(reindexRequiredUserIDs, payload.RequireReindexUserIDs()...)
	reindexRequiredUserIDs = append(reindexRequiredUserIDs, payload.DeletedUserIDs()...)

	if len(reindexRequiredUserIDs) > 0 {
		err := s.Database.WithTx(ctx, func(ctx context.Context) error {
			return s.Reindexer.MarkUsersAsReindexRequiredInTx(ctx, reindexRequiredUserIDs)
		})
		if err != nil {
			return err
		}
		for _, userID := range reindexRequiredUserIDs {
			err := s.Reindexer.EnqueueReindexUserTask(ctx, userID)
			if err != nil {
				s.Logger.WithError(err).Error("failed to enqueue reindex user task")
				return err
			}
		}
	}

	return nil
}
