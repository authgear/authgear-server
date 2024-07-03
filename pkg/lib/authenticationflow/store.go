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
		stateKey := redisFlowStateKey(s.AppID, flow.StateToken)
		ttl := Lifetime

		_, err := conn.SetEX(s.Context, flowKey, []byte(flowKey), ttl).Result()
		if err != nil {
			return err
		}

		_, err = conn.SetEX(s.Context, stateKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetFlowByStateToken(stateToken string) (*Flow, error) {
	stateKey := redisFlowStateKey(s.AppID, stateToken)
	var flow Flow
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		bytes, err := conn.Get(s.Context, stateKey).Bytes()
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
	// We do not delete the states because there are many of them.
	// Deleting the flowID is enough to make GetFlowByStateToken to return ErrFlowNotFound.
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

func (s *StoreImpl) UpdateSession(session *Session) error {
	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn *goredis.Conn) error {
		sessionKey := redisFlowSessionKey(s.AppID, session.FlowID)

		ttl := Lifetime
		_, err = conn.SetXX(s.Context, sessionKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
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

func redisFlowStateKey(appID config.AppID, stateToken string) string {
	return fmt.Sprintf("app:%s:authenticationflow_state:%s", appID, stateToken)
}

func redisFlowSessionKey(appID config.AppID, flowID string) string {
	return fmt.Sprintf("app:%s:authenticationflow_session:%s", appID, flowID)
}
