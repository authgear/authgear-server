package analytic

import (
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const (
	CumulativeUserCountType            = "cumulative.user"
	MonthlyActiveUserCountType         = "monthly.active_user"
	WeeklyActiveUserCountType          = "weekly.active_user"
	DailyActiveUserCountType           = "daily.active_user"
	DailySignupCountType               = "daily.signup"
	DailySignupPageViewCountType       = "daily.page_view.signup"
	DailySignupUniquePageViewCountType = "daily.unique_page_view.signup"
	DailyLoginPageViewCountType        = "daily.page_view.login"
	DailyLoginUniquePageViewCountType  = "daily.unique_page_view.login"
	DailySignupWithLoginIDCountType    = "daily.signup.login_id.%s"
	DailySignupWithOAuthCountType      = "daily.signup.oauth.%s"
	DailySignupAnonymouslyCountType    = "daily.signup.anonymous"
)

type DailySignupCountTypeByChannel struct {
	// ChannelName is the name of channel
	// It could be LoginIDKeyType or OAuthSSOProviderType. e.g. (email, username, google, anonymous)
	ChannelName string
	CountType   string
}

var DailySignupCountTypeByChannels = []*DailySignupCountTypeByChannel{}

func init() {
	for _, typ := range config.LoginIDKeyTypes {
		DailySignupCountTypeByChannels = append(DailySignupCountTypeByChannels, &DailySignupCountTypeByChannel{
			string(typ), fmt.Sprintf(DailySignupWithLoginIDCountType, typ),
		})
	}
	for _, typ := range config.OAuthSSOProviderTypes {
		DailySignupCountTypeByChannels = append(DailySignupCountTypeByChannels, &DailySignupCountTypeByChannel{
			string(typ), fmt.Sprintf(DailySignupWithOAuthCountType, typ),
		})
	}
	DailySignupCountTypeByChannels = append(DailySignupCountTypeByChannels, &DailySignupCountTypeByChannel{
		"anonymous", DailySignupAnonymouslyCountType,
	})
}

type Count struct {
	ID    string
	AppID string
	Count int
	Date  time.Time
	Type  string
}

func NewCount(appID string, count int, date time.Time, typ string) *Count {
	return &Count{
		ID:    uuid.New(),
		AppID: appID,
		Count: count,
		Date:  date,
		Type:  typ,
	}
}

func NewDailySignupWithLoginID(appID string, count int, date time.Time, loginIDType string) *Count {
	return &Count{
		ID:    uuid.New(),
		AppID: appID,
		Count: count,
		Date:  date,
		Type:  fmt.Sprintf(DailySignupWithLoginIDCountType, loginIDType),
	}
}

func NewDailySignupWithOAuth(appID string, count int, date time.Time, provider string) *Count {
	return &Count{
		ID:    uuid.New(),
		AppID: appID,
		Count: count,
		Date:  date,
		Type:  fmt.Sprintf(DailySignupWithOAuthCountType, provider),
	}
}
