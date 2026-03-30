package translation

import "github.com/authgear/authgear-server/pkg/api/model"

type UsageAlertTemplateVariables struct {
	Name         model.UsageName
	Action       model.UsageLimitAction
	Period       model.UsageLimitPeriod
	Quota        int
	CurrentValue int
}

type PartialTemplateVariables struct {
	// OTP
	Email string
	Phone string
	Code  string
	URL   string
	Link  string
	Host  string

	// OTP Additional context
	HasPassword bool

	// Forgot password
	Password string

	// Usage alert
	Usage *UsageAlertTemplateVariables
}

type PreparedTemplateVariables struct {
	AppName        string
	ClientID       string
	ClientName     string
	Code           string
	Email          string
	HasPassword    bool
	Host           string
	Link           string
	Password       string
	Phone          string
	State          string
	StaticAssetURL func(id string) (url string, err error)
	UILocales      string
	URL            string
	Usage          *UsageAlertTemplateVariables
	XState         string
}
