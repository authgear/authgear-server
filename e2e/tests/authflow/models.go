package tests

type TestCase struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
	Steps   []Step `yaml:"steps"`
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
