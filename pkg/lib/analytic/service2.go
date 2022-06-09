package analytic

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
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

type ReadCounterStore interface {
	GetDailyPageViewCount(
		appID config.AppID,
		pageType PageType,
		date *time.Time,
	) (pageView int, uniquePageView int, redisKeys []string, err error)
	GetDailyActiveUserCount(appID config.AppID, date *time.Time) (count int, redisKey string, err error)
	GetWeeklyActiveUserCount(appID config.AppID, year int, week int) (count int, redisKey string, err error)
	GetMonthlyActiveUserCount(appID config.AppID, year int, month int) (count int, redisKey string, err error)
	SetKeysExpire(keys []string, expiration time.Duration) error
}

type Service2 struct {
	ReadCounter ReadCounterStore
}

func (s *Service2) GetDailyCountResult(appID config.AppID, date *time.Time) (*DailyCountResult, error) {
	redisKeys := []string{}
	signupPageView, signupUniquePageView, keys, err := s.ReadCounter.GetDailyPageViewCount(appID, PageTypeSignup, date)
	if err != nil {
		return nil, err
	}
	redisKeys = append(redisKeys, keys...)

	loginPageView, loginUniquePageView, keys, err := s.ReadCounter.GetDailyPageViewCount(appID, PageTypeLogin, date)
	if err != nil {
		return nil, err
	}
	redisKeys = append(redisKeys, keys...)

	activeUserCount, dailyActiveUserKey, err := s.ReadCounter.GetDailyActiveUserCount(appID, date)
	if err != nil {
		return nil, err
	}
	redisKeys = append(redisKeys, dailyActiveUserKey)

	return &DailyCountResult{
		ActiveUser:           activeUserCount,
		SignupPageView:       signupPageView,
		SignupUniquePageView: signupUniquePageView,
		LoginPageView:        loginPageView,
		LoginUniquePageView:  loginUniquePageView,
		CountResult: &CountResult{
			RedisKeys: redisKeys,
		},
	}, nil
}

func (s *Service2) GetWeeklyCountResult(appID config.AppID, year int, week int) (*WeeklyCountResult, error) {
	activeUserCount, weeklyActiveUserKey, err := s.ReadCounter.GetWeeklyActiveUserCount(appID, year, week)
	if err != nil {
		return nil, err
	}
	return &WeeklyCountResult{
		ActiveUser: activeUserCount,
		CountResult: &CountResult{
			RedisKeys: []string{weeklyActiveUserKey},
		},
	}, nil
}

func (s *Service2) GetMonthlyCountResult(appID config.AppID, year int, month int) (*MonthlyCountResult, error) {
	activeUserCount, monthlyActiveUserKey, err := s.ReadCounter.GetMonthlyActiveUserCount(appID, year, month)
	if err != nil {
		return nil, err
	}
	return &MonthlyCountResult{
		ActiveUser: activeUserCount,
		CountResult: &CountResult{
			RedisKeys: []string{monthlyActiveUserKey},
		},
	}, nil
}

func (s *Service2) SetKeysExpire(keys []string, expiration time.Duration) error {
	return s.ReadCounter.SetKeysExpire(keys, expiration)
}
