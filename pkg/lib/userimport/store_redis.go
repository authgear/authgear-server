package userimport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

const JobTTL = 24 * time.Hour

type StoreRedis struct {
	AppID config.AppID
	Redis *appredis.Handle
}

func (s *StoreRedis) CreateJob(ctx context.Context, job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		key := redisJobKey(s.AppID, job.ID)
		ttl := JobTTL
		_, err := conn.SetEx(ctx, key, data, ttl).Result()
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreRedis) GetJob(ctx context.Context, jobID string) (*Job, error) {
	key := redisJobKey(s.AppID, jobID)
	var job Job
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		bytes, err := conn.Get(ctx, key).Bytes()
		if errors.Is(err, goredis.Nil) {
			return api.ErrTaskNotFound
		}
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &job)
		if err != nil {
			return err
		}

		return nil
	})
	return &job, err
}

func redisJobKey(appID config.AppID, jobID string) string {
	return fmt.Sprintf("app:%s:user-import:%s", appID, jobID)
}
