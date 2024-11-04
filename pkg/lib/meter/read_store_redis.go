package meter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
)

// ReadStoreRedis provides methods to get analytic counts and set expiry to the
// keys after those count are collected
type ReadStoreRedis struct {
	Redis *analyticredis.Handle
}

func (s *ReadStoreRedis) GetDailyPageViewCount(
	ctx context.Context,
	appID config.AppID,
	pageType PageType,
	date *time.Time,
) (pageView int, uniquePageView int, redisKeys []string, err error) {
	if s.Redis == nil {
		err = ErrMeterRedisIsNotConfigured
		return
	}
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		pageViewKey := dailyPageView(appID, pageType, date)
		pageView, err = s.getCount(ctx, conn, pageViewKey)
		if err != nil {
			return err
		}

		uniquePageViewKey := dailyUniquePageView(appID, pageType, date)
		uniquePageView, err = s.getPFCountWithConn(ctx, conn, uniquePageViewKey)
		if err != nil {
			return err
		}

		redisKeys = []string{
			pageViewKey,
			uniquePageViewKey,
		}
		return nil
	})
	return
}

func (s *ReadStoreRedis) GetDailyActiveUserCount(ctx context.Context, appID config.AppID, date *time.Time) (count int, redisKey string, err error) {
	redisKey = dailyActiveUserCount(appID, date)
	count, err = s.getPFCount(ctx, redisKey)
	return
}

func (s *ReadStoreRedis) GetWeeklyActiveUserCount(ctx context.Context, appID config.AppID, year int, week int) (count int, redisKey string, err error) {
	redisKey = weeklyActiveUserCount(appID, year, week)
	count, err = s.getPFCount(ctx, redisKey)
	return
}

func (s *ReadStoreRedis) GetMonthlyActiveUserCount(ctx context.Context, appID config.AppID, year int, month int) (count int, redisKey string, err error) {
	redisKey = monthlyActiveUserCount(appID, year, month)
	count, err = s.getPFCount(ctx, redisKey)
	return
}

func (s *ReadStoreRedis) SetKeysExpire(ctx context.Context, keys []string, expiration time.Duration) error {
	if s.Redis == nil {
		return nil
	}
	if len(keys) == 0 {
		return nil
	}
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		for _, key := range keys {
			_, err := conn.Expire(ctx, key, expiration).Result()
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *ReadStoreRedis) getPFCountWithConn(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string) (count int, err error) {
	result, err := conn.PFCount(ctx, key).Result()
	if err != nil {
		err = fmt.Errorf("failed to get pfcount: %w", err)
		return
	}
	count = int(result)
	return
}

func (s *ReadStoreRedis) getPFCount(ctx context.Context, key string) (count int, err error) {
	if s.Redis == nil {
		err = ErrMeterRedisIsNotConfigured
		return
	}
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		count, err = s.getPFCountWithConn(ctx, conn, key)
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (s *ReadStoreRedis) getCount(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string) (count int, err error) {
	countStr, err := conn.Get(ctx, key).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, nil
		}
		err = fmt.Errorf("failed to get count: %w", err)
		return
	}
	count, err = strconv.Atoi(countStr)
	if err != nil {
		err = fmt.Errorf("failed to parse count: %w", err)
		return
	}
	return
}
