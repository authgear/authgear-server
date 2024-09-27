package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

const Lifetime = duration.UserInteraction

type StoreImpl struct {
	Redis   *appredis.Handle
	AppID   config.AppID
	Context context.Context
}

func (s *StoreImpl) CreateWorkflow(workflow *Workflow) error {
	bytes, err := json.Marshal(workflow)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		workflowKey := redisWorkflowKey(s.AppID, workflow.WorkflowID)
		instanceKey := redisWorkflowInstanceKey(s.AppID, workflow.InstanceID)
		ttl := Lifetime

		_, err := conn.SetEX(s.Context, workflowKey, []byte(workflowKey), ttl).Result()
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

func (s *StoreImpl) GetWorkflowByInstanceID(instanceID string) (*Workflow, error) {
	instanceKey := redisWorkflowInstanceKey(s.AppID, instanceID)
	var workflow Workflow
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(s.Context, instanceKey).Bytes()
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
		_, err = conn.Get(s.Context, workflowKey).Result()
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

func (s *StoreImpl) DeleteWorkflow(workflow *Workflow) error {
	// We do not delete the instances because there are many of them.
	// Deleting the workflowID is enough to make GetWorkflowByInstanceID to return ErrWorkflowNotFound.
	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		workflowKey := redisWorkflowKey(s.AppID, workflow.WorkflowID)

		_, err := conn.Del(s.Context, workflowKey).Result()
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

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		sessionKey := redisWorkflowSessionKey(s.AppID, session.WorkflowID)
		ttl := Lifetime

		_, err := conn.SetEX(s.Context, sessionKey, bytes, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreImpl) GetSession(workflowID string) (*Session, error) {
	sessionKey := redisWorkflowSessionKey(s.AppID, workflowID)
	var session Session
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(s.Context, sessionKey).Bytes()
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

func (s *StoreImpl) DeleteSession(session *Session) error {
	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		sessionKey := redisWorkflowSessionKey(s.AppID, session.WorkflowID)

		_, err := conn.Del(s.Context, sessionKey).Result()
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
