package newinteraction

import (
	"encoding/json"
	"fmt"
	"time"

	goredis "github.com/gomodule/redigo/redis"

	"github.com/authgear/authgear-server/pkg/core/errors"
	"github.com/authgear/authgear-server/pkg/redis"
)

type StoreRedis struct {
	Redis *redis.Handle
}

func (s *StoreRedis) CreateGraph(graph *Graph) error {
	return s.create(graph, "NX")
}

func (s *StoreRedis) CreateGraphInstance(graph *Graph) error {
	return s.create(graph, "XX")
}

func (s *StoreRedis) create(graph *Graph, graphSetMode string) error {
	bytes, err := json.Marshal(graph)
	if err != nil {
		return err
	}

	return s.Redis.WithConn(func(conn redis.Conn) error {
		graphKey := redisGraphKey(graph.GraphID)
		instanceKey := redisInstanceKey(graph.InstanceID)
		ttl := toMilliseconds(GraphLifetime)
		_, err := goredis.String(conn.Do("SET", graphKey, []byte(graphKey), "PX", ttl, graphSetMode))
		if errors.Is(err, goredis.ErrNil) {
			return ErrStateNotFound
		} else if err != nil {
			return err
		}

		_, err = goredis.String(conn.Do("SET", instanceKey, bytes, "PX", ttl, "NX"))
		if errors.Is(err, goredis.ErrNil) {
			return errors.New("failed to create interaction graph instance")
		} else if err != nil {
			return err
		}

		return nil
	})
}

func (s *StoreRedis) GetGraphInstance(instanceID string) (*Graph, error) {
	instanceKey := redisInstanceKey(instanceID)
	var graph Graph
	err := s.Redis.WithConn(func(conn redis.Conn) error {
		data, err := goredis.Bytes(conn.Do("GET", instanceKey))
		if errors.Is(err, goredis.ErrNil) {
			return ErrStateNotFound
		} else if err != nil {
			return err
		}

		err = json.Unmarshal(data, &graph)
		if err != nil {
			return err
		}

		graphKey := redisGraphKey(graph.GraphID)
		_, err = goredis.String(conn.Do("GET", graphKey))
		if errors.Is(err, goredis.ErrNil) {
			return ErrStateNotFound
		} else if err != nil {
			return err
		}

		return nil
	})
	return &graph, err
}

func (s *StoreRedis) DeleteGraph(graph *Graph) error {
	return s.Redis.WithConn(func(conn redis.Conn) error {
		graphKey := redisGraphKey(graph.GraphID)
		_, err := conn.Do("DEL", graphKey)
		if err != nil {
			return err
		}
		return err
	})
}

func toMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func redisGraphKey(graphID string) string {
	return fmt.Sprintf("interaction-graph:%s", graphID)
}

func redisInstanceKey(instanceID string) string {
	return fmt.Sprintf("interaction-graph-instance:%s", instanceID)
}
