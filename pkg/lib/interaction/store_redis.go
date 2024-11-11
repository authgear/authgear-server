package interaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type StoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
}

func (s *StoreRedis) CreateGraph(ctx context.Context, graph *Graph) error {
	return s.create(ctx, graph, true)
}

func (s *StoreRedis) CreateGraphInstance(ctx context.Context, graph *Graph) error {
	return s.create(ctx, graph, false)
}

func (s *StoreRedis) create(ctx context.Context, graph *Graph, ifNotExists bool) error {
	bytes, err := json.Marshal(graph)
	if err != nil {
		return err
	}

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		graphKey := redisGraphKey(s.AppID, graph.GraphID)
		instanceKey := redisInstanceKey(s.AppID, graph.InstanceID)
		ttl := GraphLifetime

		var err error
		if ifNotExists {
			_, err = conn.SetNX(ctx, graphKey, []byte(graphKey), ttl).Result()
		} else {
			_, err = conn.SetXX(ctx, graphKey, []byte(graphKey), ttl).Result()
		}
		if errors.Is(err, goredis.Nil) {
			return ErrGraphNotFound
		} else if err != nil {
			return err
		}

		_, err = conn.SetNX(ctx, instanceKey, bytes, ttl).Result()
		if errors.Is(err, goredis.Nil) {
			return errors.New("failed to create interaction graph instance")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreRedis) GetGraphInstance(ctx context.Context, instanceID string) (*Graph, error) {
	instanceKey := redisInstanceKey(s.AppID, instanceID)
	var graph Graph
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, instanceKey).Bytes()
		if errors.Is(err, goredis.Nil) {
			return ErrGraphNotFound
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &graph)
		if err != nil {
			return err
		}

		graphKey := redisGraphKey(s.AppID, graph.GraphID)
		_, err = conn.Get(ctx, graphKey).Result()
		if errors.Is(err, goredis.Nil) {
			return ErrGraphNotFound
		} else if err != nil {
			return err
		}

		return nil
	})
	return &graph, err
}

func (s *StoreRedis) DeleteGraph(ctx context.Context, graph *Graph) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		graphKey := redisGraphKey(s.AppID, graph.GraphID)
		_, err := conn.Del(ctx, graphKey).Result()
		if err != nil {
			return err
		}
		return err
	})
}

func redisGraphKey(appID config.AppID, graphID string) string {
	return fmt.Sprintf("app:%s:interaction-graph:%s", appID, graphID)
}

func redisInstanceKey(appID config.AppID, instanceID string) string {
	return fmt.Sprintf("app:%s:interaction-graph-instance:%s", appID, instanceID)
}
