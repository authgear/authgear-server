package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const Lifetime = duration.UserInteraction

type StoreImpl struct {
	Redis   *appredis.Handle
	AppID   config.AppID
	Context context.Context
}

func (s *StoreImpl) CreateFlow(flow *Flow) error {
	bytes, err := json.Marshal(flow)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		flowKey := redisFlowKey(s.AppID, flow.FlowID)
		instanceKey := redisFlowInstanceKey(s.AppID, flow.InstanceID)
		ttl := Lifetime

		_, err := conn.SetEX(s.Context, flowKey, []byte(flowKey), ttl).Result()
		if err != nil {
			return err
		}

		_, err = conn.SetEX(s.Context, instanceKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetFlowByInstanceID(instanceID string) (*Flow, error) {
	instanceKey := redisFlowInstanceKey(s.AppID, instanceID)
	var flow Flow
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		bytes, err := conn.Get(s.Context, instanceKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrFlowNotFound
		}
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &flow)
		if err != nil {
			return err
		}

		flowKey := redisFlowKey(s.AppID, flow.FlowID)
		_, err = conn.Get(s.Context, flowKey).Result()
		if errors.Is(err, goredis.Nil) {
			return ErrFlowNotFound
		}
		if err != nil {
			return err
		}

		return nil
	})
	return &flow, err
}

func (s *StoreImpl) DeleteFlow(flow *Flow) error {
	// We do not delete the instances because there are many of them.
	// Deleting the flowID is enough to make GetFlowByInstanceID to return ErrFlowNotFound.
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		flowKey := redisFlowKey(s.AppID, flow.FlowID)

		_, err := conn.Del(s.Context, flowKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) CreateSession(session *Session) error {
	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		sessionKey := redisFlowSessionKey(s.AppID, session.FlowID)
		ttl := Lifetime

		_, err := conn.SetEX(s.Context, sessionKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetSession(flowID string) (*Session, error) {
	sessionKey := redisFlowSessionKey(s.AppID, flowID)
	var session Session
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		bytes, err := conn.Get(s.Context, sessionKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrFlowNotFound
		}
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &session)
		if err != nil {
			return err
		}

		return nil
	})
	return &session, err
}

func (s *StoreImpl) DeleteSession(session *Session) error {
	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		sessionKey := redisFlowSessionKey(s.AppID, session.FlowID)

		_, err := conn.Del(s.Context, sessionKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func redisFlowKey(appID config.AppID, flowID string) string {
	return fmt.Sprintf("app:%s:authenticationflow_flow:%s", appID, flowID)
}

func redisFlowInstanceKey(appID config.AppID, instanceID string) string {
	return fmt.Sprintf("app:%s:authenticationflow_instance:%s", appID, instanceID)
}

func redisFlowSessionKey(appID config.AppID, flowID string) string {
	return fmt.Sprintf("app:%s:authenticationflow_session:%s", appID, flowID)
}
