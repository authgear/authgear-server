package elasticsearch

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("elasticsearch-sink")} }

type Sink struct {
	Logger   Logger
	Service  Service
	Database *appdb.Handle
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
			return s.Service.MarkUsersAsReindexRequired(ctx, reindexRequiredUserIDs)
		})
		if err != nil {
			return err
		}
		for _, userID := range reindexRequiredUserIDs {
			err := s.Service.EnqueueReindexUserTask(ctx, userID)
			if err != nil {
				s.Logger.WithError(err).Error("failed to enqueue reindex user task")
				return err
			}
		}
	}

	return nil
}
