package hook

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type deliverer interface {
	WillDeliverBlockingEvent(eventType event.Type) bool
	WillDeliverNonBlockingEvent(eventType event.Type) bool
	DeliverBlockingEvent(event *event.Event) error
	DeliverNonBlockingEvent(event *event.Event) error
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("hook-sink")} }

type Sink struct {
	Logger    Logger
	Deliverer deliverer
}

func (s *Sink) ReceiveBlockingEvent(e *event.Event) (err error) {
	if s.Deliverer.WillDeliverBlockingEvent(e.Type) {
		err = s.Deliverer.DeliverBlockingEvent(e)
		if err != nil {
			if !apierrors.IsKind(err, WebHookDisallowed) {
				err = fmt.Errorf("failed to dispatch event: %w", err)
			}
			return
		}
	}

	return
}

func (s *Sink) ReceiveNonBlockingEvent(e *event.Event) (err error) {
	if s.Deliverer.WillDeliverNonBlockingEvent(e.Type) {
		if err := s.Deliverer.DeliverNonBlockingEvent(e); err != nil {
			s.Logger.WithError(err).Error("failed to dispatch non blocking event")
		}
	}

	return
}
