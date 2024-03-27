package tests

type TestCase struct {
	Name               string             `yaml:"name"`
	Path               string             `yaml:"path"`
	Focus              bool               `yaml:"focus"`
	AuthgearYAMLSource AuthgearYAMLSource `yaml:"authgear.yaml"`
	Steps              []Step             `yaml:"steps"`
	Before             []BeforeHook       `yaml:"before"`
}

func (tc *TestCase) GetFullName() string {
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
	Assert []Assert   `yaml:"assert"`
}

type AssertField string

const (
	AssertFieldActionType  AssertField = "result.action.type"
	AssertFieldErrorReason AssertField = "error.reason"
)

type AssertOp string

const (
	AssertOpEq       AssertOp = "eq"
	AssertOpNeq      AssertOp = "ne"
	AssertOpContains AssertOp = "contains"
)

type Assert struct {
	Field AssertField `yaml:"field"`
	Op    AssertOp    `yaml:"op"`
	Value string      `yaml:"value"`
}
