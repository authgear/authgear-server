package testrunner

type TestCase struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
	// Applying focus to a test case will make it the only test case to run,
	// mainly used for debugging new test cases.
	Focus              bool               `yaml:"focus"`
	AuthgearYAMLSource AuthgearYAMLSource `yaml:"authgear.yaml"`
	Steps              []Step             `yaml:"steps"`
	Before             []BeforeHook       `yaml:"before"`
}

func (tc *TestCase) FullName() string {
	return tc.Path + "/" + tc.Name
}

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
	StepActionCreate StepAction = "create"
	StepActionInput  StepAction = "input"
)

type Step struct {
	Name   string     `yaml:"name"`
	Action StepAction `yaml:"action"`
	Input  string     `yaml:"input"`
	Output *Output    `yaml:"output"`
}

type Output struct {
	Result string `yaml:"result"`
	Error  string `yaml:"error"`
}
