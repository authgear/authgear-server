package interaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
)

type StoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
}

func (s *StoreRedis) CreateGraph(graph *Graph) error {
	return s.create(graph, true)
}

func (s *StoreRedis) CreateGraphInstance(graph *Graph) error {
	return s.create(graph, false)
}

func (s *StoreRedis) create(graph *Graph, ifNotExists bool) error {
	ctx := context.Background()
	bytes, err := json.Marshal(graph)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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

func (s *StoreRedis) GetGraphInstance(instanceID string) (*Graph, error) {
	ctx := context.Background()
	instanceKey := redisInstanceKey(s.AppID, instanceID)
	var graph Graph
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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

func (s *StoreRedis) DeleteGraph(graph *Graph) error {
	ctx := context.Background()
	return s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
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
