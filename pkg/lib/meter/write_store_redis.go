package meter

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type PageType string

const (
	PageTypeSignup PageType = "signup"
	PageTypeLogin  PageType = "login"
)

type StoreRedisLogger struct{ *log.Logger }

func NewStoreRedisLogger(lf *log.Factory) StoreRedisLogger {
	return StoreRedisLogger{lf.New("redis-analytic-store")}
}

type WriteStoreRedis struct {
	Context context.Context
	Redis   *analyticredis.Handle
	AppID   config.AppID
	Clock   clock.Clock
	Logger  StoreRedisLogger
}

func (s *WriteStoreRedis) TrackActiveUser(userID string) (err error) {
	if s.Redis == nil {
		return nil
	}
	now := s.Clock.NowUTC()
	year, week := now.ISOWeek()
	month := now.Month()
	keys := []string{
		monthlyActiveUserCount(s.AppID, year, int(month)),
		weeklyActiveUserCount(s.AppID, year, week),
		dailyActiveUserCount(s.AppID, &now),
	}
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		for _, key := range keys {
			_, err := conn.PFAdd(s.Context, key, userID).Result()
			if err != nil {
				err = fmt.Errorf("failed to track user active count: %w", err)
				return err
			}
		}
		return nil
	})
	return
}

func (s *WriteStoreRedis) TrackPageView(visitorID string, pageType PageType) (err error) {
	if s.Redis == nil {
		return nil
	}
	now := s.Clock.NowUTC()
	err = s.Redis.WithConn(func(conn redis.Redis_6_0_Cmdable) error {
		uniquePageViewKey := dailyUniquePageView(s.AppID, pageType, &now)
		_, err := conn.PFAdd(s.Context, uniquePageViewKey, visitorID).Result()
		if err != nil {
			err = fmt.Errorf("failed to track unique page view: %w", err)
			return err
		}

		pageViewKey := dailyPageView(s.AppID, pageType, &now)
		_, err = conn.Incr(s.Context, pageViewKey).Result()
		if err != nil {
			err = fmt.Errorf("failed to track page view: %w", err)
			return err
		}

		return nil
	})
	return
}

func monthlyActiveUserCount(appID config.AppID, year int, month int) string {
	return fmt.Sprintf("app:%s:monthly-active-user:%04d-%02d", appID, year, month)
}

func weeklyActiveUserCount(appID config.AppID, year int, week int) string {
	return fmt.Sprintf("app:%s:weekly-active-user:%04d-W%02d", appID, year, week)
}

func dailyActiveUserCount(appID config.AppID, date *time.Time) string {
	return fmt.Sprintf("app:%s:daily-active-user:%s", appID, date.Format(timeutil.LayoutISODate))
}

func dailyUniquePageView(appID config.AppID, page PageType, date *time.Time) string {
	return fmt.Sprintf("app:%s:daily-unique-page-view:%s:%s", appID, page, date.Format(timeutil.LayoutISODate))
}

func dailyPageView(appID config.AppID, page PageType, date *time.Time) string {
	return fmt.Sprintf("app:%s:daily-page-view:%s:%s", appID, page, date.Format(timeutil.LayoutISODate))
}
