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
	return nil
}
