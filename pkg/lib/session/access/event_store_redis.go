package access

import (
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

const maxEventStreamLength = 10

const eventTypeAccessEvent = "access"

type EventStoreRedis struct {
	Redis *redis.Handle
	AppID config.AppID
}

func (s *EventStoreRedis) AppendEvent(sessionID string, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	streamKey := accessEventStreamKey(s.AppID, sessionID)
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

func (s *EventStoreRedis) ResetEventStream(sessionID string) error {
	streamKey := accessEventStreamKey(s.AppID, sessionID)

	return s.Redis.WithConn(func(conn redis.Conn) error {
		_, err := conn.Do("DEL", streamKey)
		if err != nil {
			return err
		}

		return nil
	})
}

func accessEventStreamKey(appID config.AppID, sessionID string) string {
	return fmt.Sprintf("%s:access-events:%s", appID, sessionID)
}
