package redis

import (
	"context"
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/redis"
)

// TODO(session): tune event persistence, maybe use other datastore
const maxEventStreamLength = 10

const eventTypeAccessEvent = "access"

type EventStore struct {
	ctx   context.Context
	appID string
}

var _ session.EventStore = &EventStore{}

func NewEventStore(ctx context.Context, appID string) *EventStore {
	return &EventStore{ctx: ctx, appID: appID}
}

func (s *EventStore) AppendAccessEvent(session *session.Session, event *session.AccessEvent) (err error) {
	json, err := json.Marshal(event)
	if err != nil {
		return
	}

	conn := redis.GetConn(s.ctx)
	key := eventStreamKey(s.appID, session.ID)

	args := []interface{}{key}
	if maxEventStreamLength >= 0 {
		args = append(args, "MAXLEN", "~", maxEventStreamLength)
	}
	args = append(args, "*", eventTypeAccessEvent, json)

	_, err = conn.Do("XADD", args...)
	return
}
