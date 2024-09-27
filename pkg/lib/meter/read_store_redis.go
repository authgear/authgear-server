package meter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
)

// ReadStoreRedis provides methods to get analytic counts and set expiry to the
// keys after those count are collected
type ReadStoreRedis struct {
	Context context.Context
	Redis   *analyticredis.Handle
}

func (s *ReadStoreRedis) GetDailyPageViewCount(
	appID config.AppID,
	pageType PageType,
	date *time.Time,
) (pageView int, uniquePageView int, redisKeys []string, err error) {
	if s.Redis == nil {
		err = ErrMeterRedisIsNotConfigured
		return
	}
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		pageViewKey := dailyPageView(appID, pageType, date)
		pageView, err = s.getCount(conn, pageViewKey)
		if err != nil {
			return err
		}

		uniquePageViewKey := dailyUniquePageView(appID, pageType, date)
		uniquePageView, err = s.getPFCountWithConn(conn, uniquePageViewKey)
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

func (s *ReadStoreRedis) GetDailyActiveUserCount(appID config.AppID, date *time.Time) (count int, redisKey string, err error) {
	redisKey = dailyActiveUserCount(appID, date)
	count, err = s.getPFCount(redisKey)
	return
}

func (s *ReadStoreRedis) GetWeeklyActiveUserCount(appID config.AppID, year int, week int) (count int, redisKey string, err error) {
	redisKey = weeklyActiveUserCount(appID, year, week)
	count, err = s.getPFCount(redisKey)
	return
}

func (s *ReadStoreRedis) GetMonthlyActiveUserCount(appID config.AppID, year int, month int) (count int, redisKey string, err error) {
	redisKey = monthlyActiveUserCount(appID, year, month)
	count, err = s.getPFCount(redisKey)
	return
}

func (s *ReadStoreRedis) SetKeysExpire(keys []string, expiration time.Duration) error {
	if s.Redis == nil {
		return nil
	}
	if len(keys) == 0 {
		return nil
	}
	err := s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		for _, key := range keys {
			_, err := conn.Expire(s.Context, key, expiration).Result()
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

func (s *ReadStoreRedis) getPFCountWithConn(conn redis.Redis_6_0_Cmdable, key string) (count int, err error) {
	result, err := conn.PFCount(s.Context, key).Result()
	if err != nil {
		err = fmt.Errorf("failed to get pfcount: %w", err)
		return
	}
	count = int(result)
	return
}

func (s *ReadStoreRedis) getPFCount(key string) (count int, err error) {
	if s.Redis == nil {
		err = ErrMeterRedisIsNotConfigured
		return
	}
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		count, err = s.getPFCountWithConn(conn, key)
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (s *ReadStoreRedis) getCount(conn redis.Redis_6_0_Cmdable, key string) (count int, err error) {
	countStr, err := conn.Get(s.Context, key).Result()
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
