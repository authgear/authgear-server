package translation

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
	AppName  string
	Password string
}

type PreparedTemplateVariables struct {
	AppName        string
	ClientID       string
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
	XState         string
}
