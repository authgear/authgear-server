package testrunner

type AuthgearYAMLSource struct {
	Extend   string `yaml:"extend"`
	Override string `yaml:"override"`
}

type BeforeHookType string

const (
	BeforeHookTypeUserImport BeforeHookType = "user_import"
	BeforeHookTypeCustomSQL  BeforeHookType = "custom_sql"
)

type BeforeHookCustomSQL struct {
	Path string `yaml:"path"`
}

type BeforeHook struct {
	Type       BeforeHookType      `yaml:"type"`
	UserImport string              `yaml:"user_import"`
	CustomSQL  BeforeHookCustomSQL `yaml:"custom_sql"`
}

type StepAction string

const (
	StepActionCreate        StepAction = "create"
	StepActionInput         StepAction = "input"
	StepActionOAuthRedirect StepAction = "oauth_redirect"
)

type Step struct {
	Name   string     `yaml:"name"`
	Action StepAction `yaml:"action"`

	// `action` == "create" or "input"
	Input string `yaml:"input"`

	// `action` == "oauth_redirect"
	To          string `yaml:"to"`
	RedirectURI string `yaml:"redirect_uri"`

	Output *Output `yaml:"output"`
}

type Output struct {
	Result string `yaml:"result"`
	Error  string `yaml:"error"`
}

type StepResult struct {
	Result interface{}
	Error  error
}
