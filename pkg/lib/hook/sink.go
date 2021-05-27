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

type store interface {
	NextSequenceNumber() (int64, error)
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("hook-sink")} }

type Sink struct {
	Logger    Logger
	Store     store
	Deliverer deliverer
}

func (s *Sink) ReceiveBlockingEvent(e *event.Event) (err error) {
	var seq int64

	if s.Deliverer.WillDeliverBlockingEvent(e.Type) {
		seq, err = s.Store.NextSequenceNumber()
		if err != nil {
			err = fmt.Errorf("failed to dispatch event: %w", err)
			return
		}
		e.SetSeq(seq)

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
	var seq int64

	if s.Deliverer.WillDeliverNonBlockingEvent(e.Type) {
		seq, err = s.Store.NextSequenceNumber()
		if err != nil {
			err = fmt.Errorf("failed to persist event: %w", err)
			return
		}
		e.SetSeq(seq)

		if err := s.Deliverer.DeliverNonBlockingEvent(e); err != nil {
			s.Logger.WithError(err).Error("failed to dispatch non blocking event")
		}
	}

	return
}
