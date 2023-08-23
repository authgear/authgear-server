package authenticationflow

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/pubsub"
	goredis "github.com/go-redis/redis/v8"
)

type eventRedisPool struct{ *appredis.Handle }

func (p eventRedisPool) Get() *goredis.Client {
	return p.Handle.Client()
}

type EventStoreImpl struct {
	AppID       config.AppID
	RedisHandle *appredis.Handle
	Store       Store
	publisher   *pubsub.Publisher
}

func NewEventStore(appID config.AppID, handle *appredis.Handle, store Store) *EventStoreImpl {
	s := &EventStoreImpl{
		AppID:       appID,
		RedisHandle: handle,
		Store:       store,
		publisher:   &pubsub.Publisher{RedisPool: eventRedisPool{handle}},
	}
	return s
}

func (s *EventStoreImpl) Publish(flowID string, e Event) error {
	channelName, err := s.ChannelName(flowID)
	if errors.Is(err, ErrFlowNotFound) {
		// Treat events to an non-existent (e.g. expired) flow as noop.
		return nil
	} else if err != nil {
		return err
	}

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	err = s.publisher.Publish(channelName, b)
	if err != nil {
		return err
	}

	return nil
}

func (s *EventStoreImpl) ChannelName(flowID string) (string, error) {
	// Ignore events for flows without session.
	_, err := s.Store.GetSession(flowID)
	if err != nil {
		return "", err
	}

	channelName := fmt.Sprintf("app:%s:authenticationflow-events:%s", s.AppID, flowID)
	return channelName, nil
}
