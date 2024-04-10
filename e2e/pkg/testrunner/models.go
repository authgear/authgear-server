package testrunner

import authflowclient "github.com/authgear/authgear-server/e2e/pkg/e2eclient"

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

type StepResult struct {
	Result *authflowclient.FlowResponse
	Error  error
}
