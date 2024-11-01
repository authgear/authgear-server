package authenticationflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const Lifetime = duration.UserInteraction

type StoreImpl struct {
	Redis *appredis.Handle
	AppID config.AppID
}

func (s *StoreImpl) CreateFlow(ctx context.Context, flow *Flow) error {
	bytes, err := json.Marshal(flow)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		flowKey := redisFlowKey(s.AppID, flow.FlowID)
		stateKey := redisFlowStateKey(s.AppID, flow.StateToken)
		ttl := Lifetime

		_, err := conn.SetEx(ctx, flowKey, []byte(flowKey), ttl).Result()
		if err != nil {
			return err
		}

		_, err = conn.SetEx(ctx, stateKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetFlowByStateToken(ctx context.Context, stateToken string) (*Flow, error) {
	stateKey := redisFlowStateKey(s.AppID, stateToken)
	var flow Flow
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(ctx, stateKey).Bytes()
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
		_, err = conn.Get(ctx, flowKey).Result()
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

func (s *StoreImpl) DeleteFlow(ctx context.Context, flow *Flow) error {
	// We do not delete the states because there are many of them.
	// Deleting the flowID is enough to make GetFlowByStateToken to return ErrFlowNotFound.
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		flowKey := redisFlowKey(s.AppID, flow.FlowID)

		_, err := conn.Del(ctx, flowKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) CreateSession(ctx context.Context, session *Session) error {
	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		sessionKey := redisFlowSessionKey(s.AppID, session.FlowID)
		ttl := Lifetime

		_, err := conn.SetEx(ctx, sessionKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetSession(ctx context.Context, flowID string) (*Session, error) {
	sessionKey := redisFlowSessionKey(s.AppID, flowID)
	var session Session
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(ctx, sessionKey).Bytes()
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

func (s *StoreImpl) UpdateSession(ctx context.Context, session *Session) error {
	bytes, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		sessionKey := redisFlowSessionKey(s.AppID, session.FlowID)

		ttl := Lifetime
		_, err = conn.SetXX(ctx, sessionKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) DeleteSession(ctx context.Context, session *Session) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		sessionKey := redisFlowSessionKey(s.AppID, session.FlowID)

		_, err := conn.Del(ctx, sessionKey).Result()
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
