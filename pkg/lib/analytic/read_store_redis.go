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

type DailyCountResult struct {
	ActiveUser           int
	SignupPageView       int
	SignupUniquePageView int
	LoginPageView        int
	LoginUniquePageView  int
}

type WeeklyCountResult struct {
	ActiveUser int
}

type MonthlyCountResult struct {
	ActiveUser int
}
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
		signupPageView, signupUniquePageView, err := s.getDailyPageViewCount(conn, appID, PageTypeSignup, date)
		if err != nil {
			return err
		}

		loginPageView, loginUniquePageView, err := s.getDailyPageViewCount(conn, appID, PageTypeLogin, date)
		if err != nil {
			return err
		}

		activeUserCount, err := s.getPFCount(conn, dailyActiveUserCount(appID, date))
		if err != nil {
			return err
		}

		result = &DailyCountResult{
			ActiveUser:           activeUserCount,
			SignupPageView:       signupPageView,
			SignupUniquePageView: signupUniquePageView,
			LoginPageView:        loginPageView,
			LoginUniquePageView:  loginUniquePageView,
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
		activeUserCount, err := s.getPFCount(conn, weeklyActiveUserCount(appID, year, week))
		if err != nil {
			return err
		}
		result = &WeeklyCountResult{
			ActiveUser: activeUserCount,
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
		activeUserCount, err := s.getPFCount(conn, monthlyActiveUserCount(appID, year, month))
		if err != nil {
			return err
		}
		result = &MonthlyCountResult{
			ActiveUser: activeUserCount,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *ReadStoreRedis) getDailyPageViewCount(
	conn *goredis.Conn,
	appID config.AppID,
	pageType PageType,
	date *time.Time,
) (pageView int, uniquePageView int, err error) {
	pageView, err = s.getCount(conn, dailyPageView(appID, pageType, date))
	if err != nil {
		return
	}

	uniquePageView, err = s.getPFCount(conn, dailyUniquePageView(appID, pageType, date))
	if err != nil {
		return
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
