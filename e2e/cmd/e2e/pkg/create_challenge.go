package e2e

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
	"github.com/authgear/authgear-server/pkg/lib/config"
	infraredis "github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func (c *End2End) CreateChallenge(
	ctx context.Context,
	appID string,
	purpose challenge.Purpose,
	token string) (err error) {
	cfg, err := LoadConfigFromEnv()
	if err != nil {
		return err
	}
	lf := log.NewFactory(log.LevelInfo)

	redisPool := infraredis.NewPool()
	redisHub := infraredis.NewHub(ctx, redisPool, lf)
	redis := appredis.NewHandle(
		redisPool,
		redisHub,
		&cfg.RedisConfig,
		&config.RedisCredentials{
			RedisURL: cfg.GlobalRedis.RedisURL,
		},
		lf,
	)
	store := &challenge.Store{
		AppID: config.AppID(appID),
		Redis: redis,
	}
	ttl, err := time.ParseDuration("1h")
	if err != nil {
		panic(err)
	}
	clk := clock.NewSystemClock()
	now := clk.NowUTC()
	return store.Save(ctx, &challenge.Challenge{
		Token:     token,
		Purpose:   purpose,
		CreatedAt: now,
		ExpireAt:  now.Add(ttl),
	}, ttl)
}
