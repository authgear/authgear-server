package workflow

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

func (s *StoreImpl) CreateWorkflow(ctx context.Context, workflow *Workflow) error {
	bytes, err := json.Marshal(workflow)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		workflowKey := redisWorkflowKey(s.AppID, workflow.WorkflowID)
		instanceKey := redisWorkflowInstanceKey(s.AppID, workflow.InstanceID)
		ttl := Lifetime

		_, err := conn.SetEx(ctx, workflowKey, []byte(workflowKey), ttl).Result()
		if err != nil {
			return err
		}

		_, err = conn.SetEx(ctx, instanceKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetWorkflowByInstanceID(ctx context.Context, instanceID string) (*Workflow, error) {
	instanceKey := redisWorkflowInstanceKey(s.AppID, instanceID)
	var workflow Workflow
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(ctx, instanceKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrWorkflowNotFound
		}
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &workflow)
		if err != nil {
			return err
		}

		workflowKey := redisWorkflowKey(s.AppID, workflow.WorkflowID)
		_, err = conn.Get(ctx, workflowKey).Result()
		if errors.Is(err, goredis.Nil) {
			return ErrWorkflowNotFound
		}
		if err != nil {
			return err
		}

		return nil
	})
	return &workflow, err
}

func (s *StoreImpl) DeleteWorkflow(ctx context.Context, workflow *Workflow) error {
	// We do not delete the instances because there are many of them.
	// Deleting the workflowID is enough to make GetWorkflowByInstanceID to return ErrWorkflowNotFound.
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		workflowKey := redisWorkflowKey(s.AppID, workflow.WorkflowID)

		_, err := conn.Del(ctx, workflowKey).Result()
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
		sessionKey := redisWorkflowSessionKey(s.AppID, session.WorkflowID)
		ttl := Lifetime

		_, err := conn.SetEx(ctx, sessionKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetSession(ctx context.Context, workflowID string) (*Session, error) {
	sessionKey := redisWorkflowSessionKey(s.AppID, workflowID)
	var session Session
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(ctx, sessionKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrWorkflowNotFound
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

func (s *StoreImpl) DeleteSession(ctx context.Context, session *Session) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		sessionKey := redisWorkflowSessionKey(s.AppID, session.WorkflowID)

		_, err := conn.Del(ctx, sessionKey).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func redisWorkflowKey(appID config.AppID, workflowID string) string {
	return fmt.Sprintf("app:%s:workflow:%s", appID, workflowID)
}

func redisWorkflowInstanceKey(appID config.AppID, instanceID string) string {
	return fmt.Sprintf("app:%s:workflow-instance:%s", appID, instanceID)
}

func redisWorkflowSessionKey(appID config.AppID, workflowID string) string {
	return fmt.Sprintf("app:%s:workflow-session:%s", appID, workflowID)
}
