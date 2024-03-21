package elasticsearch

import (
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

func (s *Sink) ReceiveBlockingEvent(e *event.Event) error {
	return nil
}

func (s *Sink) ReceiveNonBlockingEvent(e *event.Event) error {
	payload := e.Payload.(event.NonBlockingPayload)
	reindexRequiredUserIDs := payload.RequireReindexUserIDs()
	if len(reindexRequiredUserIDs) > 0 {
		for _, userID := range reindexRequiredUserIDs {
			err := s.Database.ReadOnly(func() error {
				return s.Service.ReindexUser(userID, payload.IsUserDeleted())
			})
			if err != nil {
				s.Logger.WithError(err).Error("failed to reindex user")
				return err
			}
		}
	}
	return nil
}
