package meter

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type PageType string

const (
	PageTypeSignup PageType = "signup"
	PageTypeLogin  PageType = "login"
)

type WriteStoreRedis struct {
	Redis *analyticredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

func (s *WriteStoreRedis) TrackActiveUser(ctx context.Context, userID string) (err error) {
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
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		for _, key := range keys {
			_, err := conn.PFAdd(ctx, key, userID).Result()
			if err != nil {
				err = fmt.Errorf("failed to track user active count: %w", err)
				return err
			}
		}
		return nil
	})
	return
}

func (s *WriteStoreRedis) TrackPageView(ctx context.Context, visitorID string, pageType PageType) (err error) {
	if s.Redis == nil {
		return nil
	}
	now := s.Clock.NowUTC()
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		uniquePageViewKey := dailyUniquePageView(s.AppID, pageType, &now)
		_, err := conn.PFAdd(ctx, uniquePageViewKey, visitorID).Result()
		if err != nil {
			err = fmt.Errorf("failed to track unique page view: %w", err)
			return err
		}

		pageViewKey := dailyPageView(s.AppID, pageType, &now)
		_, err = conn.Incr(ctx, pageViewKey).Result()
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
