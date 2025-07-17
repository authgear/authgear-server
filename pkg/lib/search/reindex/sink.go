package reindex

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var SinkLogger = slogutil.NewLogger("search-reindex-sink")

type Sink struct {
	Reindexer *Reindexer
	Database  *appdb.Handle
}

func (s *Sink) ReceiveBlockingEvent(ctx context.Context, e *event.Event) error {
	return nil
}

func (s *Sink) ReceiveNonBlockingEvent(ctx context.Context, e *event.Event) error {
	logger := SinkLogger.GetLogger(ctx)
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
				logger.WithError(err).Error(ctx, "failed to enqueue reindex user task")
				return err
			}
		}
	}

	return nil
}
