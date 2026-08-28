package oauthclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

const (
	// A DCR client is immutable until RFC 7592 lands, so its only mutation is
	// deletion, which invalidates explicitly (see Commands.DeleteClient).
	// This TTL is purely the backstop for an invalidation that failed to run.
	//
	// CIMD rows ARE mutable (a refetch overwrites in place), so the CIMD
	// upsert path must invalidate too.
	dynamicClientCacheTTL = 5 * time.Minute
	// Negative entries expire faster: they exist to blunt read amplification,
	// not to be authoritative.
	dynamicClientCacheNotFoundTTL = 30 * time.Second
)

func redisKeyDynamicClient(appID string, clientID string) string {
	// The key is keyed by client_id alone, not by source: a client_id belongs
	// to exactly one source, and the resolver looks up by client_id without
	// knowing the source yet.
	//
	// clientID is caller-influenced for CIMD (it is a URL), so it must be
	// hashed rather than interpolated raw — a ':' in a URL would otherwise
	// let one client_id's key collide with another's namespace.
	return fmt.Sprintf("app:%s:dynamic-client:%s", appID, crypto.SHA256String(clientID))
}

// cachedClient is the Redis payload. Found distinguishes a cached negative
// result (Found: true, Client: nil) from "not cached at all", which Get
// reports via its own found return value instead.
type cachedClient struct {
	Found  bool    `json:"found"`
	Client *Client `json:"client,omitempty"`
}

type ClientCache struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

// Get reports found=false when clientID has no cache entry at all (a true
// cache miss). found=true with client==nil means a cached negative result.
func (c *ClientCache) Get(ctx context.Context, clientID string) (client *Client, found bool, err error) {
	err = c.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		data, err := conn.Get(ctx, redisKeyDynamicClient(string(c.AppID), clientID)).Bytes()
		if errors.Is(err, goredis.Nil) {
			found = false
			return nil
		} else if err != nil {
			return err
		}

		var cached cachedClient
		if err := json.Unmarshal(data, &cached); err != nil {
			return err
		}
		found = true
		client = cached.Client
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return client, found, nil
}

func (c *ClientCache) Set(ctx context.Context, client *Client) error {
	return c.set(ctx, client.ClientID, &cachedClient{Found: true, Client: client}, dynamicClientCacheTTL)
}

func (c *ClientCache) SetNotFound(ctx context.Context, clientID string) error {
	return c.set(ctx, clientID, &cachedClient{Found: false}, dynamicClientCacheNotFoundTTL)
}

func (c *ClientCache) set(ctx context.Context, clientID string, payload *cachedClient, ttl time.Duration) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Set(ctx, redisKeyDynamicClient(string(c.AppID), clientID), data, ttl).Result()
		return err
	})
}

func (c *ClientCache) Delete(ctx context.Context, clientID string) error {
	return c.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		_, err := conn.Del(ctx, redisKeyDynamicClient(string(c.AppID), clientID)).Result()
		return err
	})
}
