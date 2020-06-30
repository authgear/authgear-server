package redis

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/auth"
	"github.com/authgear/authgear-server/pkg/redis"
)

// TODO(session): tune event persistence, maybe use other datastore
const maxEventStreamLength = 10

const eventTypeAccessEvent = "access"

type EventStore struct {
	Redis *redis.Context
	AppID config.AppID
}

var _ auth.AccessEventStore = &EventStore{}

func (s *EventStore) AppendAccessEvent(session auth.AuthSession, event *auth.AccessEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	streamKey := accessEventStreamKey(s.AppID, session.SessionID())
	args := []interface{}{streamKey}
	if maxEventStreamLength >= 0 {
		args = append(args, "MAXLEN", "~", maxEventStreamLength)
	}
	args = append(args, "*", eventTypeAccessEvent, data)

	return s.Redis.WithConn(func(conn redis.Conn) error {
		_, err = conn.Do("XADD", args...)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *EventStore) ResetEventStream(session auth.AuthSession) error {
	streamKey := accessEventStreamKey(s.AppID, session.SessionID())

	return s.Redis.WithConn(func(conn redis.Conn) error {
		_, err := conn.Do("DEL", streamKey)
		if err != nil {
			return err
		}

		return nil
	})
}
