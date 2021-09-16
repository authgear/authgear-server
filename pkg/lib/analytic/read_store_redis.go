package analytic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
)

// CountResult includes the redis keys of the report
// Expiration should be set to those keys after storing the count to the db
type CountResult struct {
	RedisKeys []string
}

type DailyCountResult struct {
	*CountResult
	ActiveUser           int
	SignupPageView       int
	SignupUniquePageView int
	LoginPageView        int
	LoginUniquePageView  int
}

type WeeklyCountResult struct {
	*CountResult
	ActiveUser int
}

type MonthlyCountResult struct {
	*CountResult
	ActiveUser int
}

// ReadStoreRedis provides methods to get analytic counts and set expiry to the
// keys after those count are collected
type ReadStoreRedis struct {
	Context context.Context
	Redis   *analyticredis.Handle
}

func (s *ReadStoreRedis) GetDailyCountResult(appID config.AppID, date *time.Time) (*DailyCountResult, error) {
	var result *DailyCountResult
	if s.Redis == nil {
		// redis is not configured, give empty result
		result = &DailyCountResult{}
		return result, nil
	}
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		redisKeys := []string{}
		signupPageView, signupUniquePageView, keys, err := s.getDailyPageViewCount(conn, appID, PageTypeSignup, date)
		if err != nil {
			return err
		}
		redisKeys = append(redisKeys, keys...)

		loginPageView, loginUniquePageView, keys, err := s.getDailyPageViewCount(conn, appID, PageTypeLogin, date)
		if err != nil {
			return err
		}
		redisKeys = append(redisKeys, keys...)

		dailyActiveUserKey := dailyActiveUserCount(appID, date)
		activeUserCount, err := s.getPFCount(conn, dailyActiveUserKey)
		if err != nil {
			return err
		}
		redisKeys = append(redisKeys, dailyActiveUserKey)

		result = &DailyCountResult{
			ActiveUser:           activeUserCount,
			SignupPageView:       signupPageView,
			SignupUniquePageView: signupUniquePageView,
			LoginPageView:        loginPageView,
			LoginUniquePageView:  loginUniquePageView,
			CountResult: &CountResult{
				RedisKeys: redisKeys,
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ReadStoreRedis) GetWeeklyCountResult(appID config.AppID, year int, week int) (*WeeklyCountResult, error) {
	var result *WeeklyCountResult
	if s.Redis == nil {
		// redis is not configured, give empty result
		result = &WeeklyCountResult{}
		return result, nil
	}
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		weeklyActiveUserKey := weeklyActiveUserCount(appID, year, week)
		activeUserCount, err := s.getPFCount(conn, weeklyActiveUserKey)
		if err != nil {
			return err
		}
		result = &WeeklyCountResult{
			ActiveUser: activeUserCount,
			CountResult: &CountResult{
				RedisKeys: []string{weeklyActiveUserKey},
			},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ReadStoreRedis) GetMonthlyCountResult(appID config.AppID, year int, month int) (*MonthlyCountResult, error) {
	var result *MonthlyCountResult
	if s.Redis == nil {
		// redis is not configured, give empty result
		result = &MonthlyCountResult{}
		return result, nil
	}
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
		monthlyActiveUserKey := monthlyActiveUserCount(appID, year, month)
		activeUserCount, err := s.getPFCount(conn, monthlyActiveUserKey)
		if err != nil {
			return err
		}
		result = &MonthlyCountResult{
			ActiveUser: activeUserCount,
			CountResult: &CountResult{
				RedisKeys: []string{monthlyActiveUserKey},
			},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ReadStoreRedis) SetKeysExpire(keys []string, expiration time.Duration) error {
	if s.Redis == nil {
		return nil
	}
	if len(keys) == 0 {
		return nil
	}
	err := s.Redis.WithConn(func(conn *goredis.Conn) error {
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

func (s *ReadStoreRedis) getDailyPageViewCount(
	conn *goredis.Conn,
	appID config.AppID,
	pageType PageType,
	date *time.Time,
) (pageView int, uniquePageView int, redisKeys []string, err error) {
	pageViewKey := dailyPageView(appID, pageType, date)
	pageView, err = s.getCount(conn, pageViewKey)
	if err != nil {
		return
	}

	uniquePageViewKey := dailyUniquePageView(appID, pageType, date)
	uniquePageView, err = s.getPFCount(conn, uniquePageViewKey)
	if err != nil {
		return
	}

	redisKeys = []string{
		pageViewKey,
		uniquePageViewKey,
	}

	return
}

func (s *ReadStoreRedis) getPFCount(conn *goredis.Conn, key string) (count int, err error) {
	result, err := conn.PFCount(s.Context, key).Result()
	if err != nil {
		err = fmt.Errorf("failed to get pfcount: %w", err)
		return
	}
	count = int(result)
	return
}

func (s *ReadStoreRedis) getCount(conn *goredis.Conn, key string) (count int, err error) {
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
