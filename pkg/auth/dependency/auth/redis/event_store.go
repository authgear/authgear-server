package redis

import (
	"context"
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/core/redis"
)

// TODO(session): tune event persistence, maybe use other datastore
const maxEventStreamLength = 10

const eventTypeAccessEvent = "access"

type EventStore struct {
	ctx   context.Context
	appID string
}

var _ auth.AccessEventStore = &EventStore{}

func NewEventStore(ctx context.Context, appID string) *EventStore {
	return &EventStore{ctx: ctx, appID: appID}
}

func (s *EventStore) AppendAccessEvent(session auth.AuthSession, event *auth.AccessEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	conn := redis.GetConn(s.ctx)
	streamKey := accessEventStreamKey(s.appID, session.SessionID())

	args := []interface{}{streamKey}
	if maxEventStreamLength >= 0 {
		args = append(args, "MAXLEN", "~", maxEventStreamLength)
	}
	args = append(args, "*", eventTypeAccessEvent, data)

	_, err = conn.Do("XADD", args...)
	if err != nil {
		return err
	}

	return nil
}
