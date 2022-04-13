package tutorial

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type Sink struct {
	AppID        config.AppID
	Service      *Service
	GlobalHandle *globaldb.Handle
}

func (s *Sink) ReceiveBlockingEvent(e *event.Event) error {
	return nil
}

func (s *Sink) ReceiveNonBlockingEvent(e *event.Event) error {
	appID := string(s.AppID)
	if appID == "" {
		return nil
	}
	if e.Type == nonblocking.UserCreated {
		return s.GlobalHandle.WithTx(func() error {
			return s.Service.RecordProgresses(appID, []Progress{ProgressAuthui})
		})
	}

	return nil
}
