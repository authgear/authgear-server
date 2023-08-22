package workflowconfig

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaLoginFlowStepAuthenticate struct {
	Candidates         []UseAuthenticationCandidate
	DeviceTokenEnabled bool
}

var _ workflow.InputSchema = &InputSchemaLoginFlowStepAuthenticate{}

func (i *InputSchemaLoginFlowStepAuthenticate) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for index, candidate := range i.Candidates {
		index := index
		candidate := candidate

		b := validation.SchemaBuilder{}
		required := []string{"authentication"}
		b.Properties().Property("authentication", validation.SchemaBuilder{}.Const(candidate.Authentication))

		requireString := func(key string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString))
		}
		requireIndex := func() {
			required = append(required, "index")
			b.Properties().Property("index", validation.SchemaBuilder{}.
				Type(validation.TypeInteger).
				Const(index),
			)
		}

		switch candidate.Authentication {
		case config.WorkflowAuthenticationMethodPrimaryPassword:
			requireString("password")
		case config.WorkflowAuthenticationMethodSecondaryPassword:
			requireString("password")
		case config.WorkflowAuthenticationMethodSecondaryTOTP:
			requireString("code")
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
			requireIndex()
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
			requireIndex()
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
			requireIndex()
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
			requireIndex()
		case config.WorkflowAuthenticationMethodRecoveryCode:
			requireString("recovery_code")
		default:
			// Skip the following code.
			continue
		}

		b.Required(required...)
		oneOf = append(oneOf, b)
	}

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	if len(oneOf) > 0 {
		b.OneOf(oneOf...)
	}

	deviceToken := validation.SchemaBuilder{}.
		Type(validation.TypeBoolean)
	if !i.DeviceTokenEnabled {
		deviceToken.Const(false)
	}
	b.Properties().Property("request_device_token", deviceToken)

	return b
}

func (i *InputSchemaLoginFlowStepAuthenticate) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputLoginFlowStepAuthenticate
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputLoginFlowStepAuthenticate struct {
	Authentication     config.WorkflowAuthenticationMethod `json:"authentication,omitempty"`
	RequestDeviceToken bool                                `json:"request_device_token,omitempty"`
	Password           string                              `json:"password,omitempty"`
	Code               string                              `json:"code,omitempty"`
	RecoveryCode       string                              `json:"recovery_code,omitempty"`
	Index              int                                 `json:"index,omitempty"`
}

var _ workflow.Input = &InputLoginFlowStepAuthenticate{}
var _ inputTakeAuthenticationMethod = &InputLoginFlowStepAuthenticate{}
var _ inputDeviceTokenRequested = &InputLoginFlowStepAuthenticate{}
var _ inputTakePassword = &InputLoginFlowStepAuthenticate{}
var _ inputTakeTOTP = &InputLoginFlowStepAuthenticate{}
var _ inputTakeRecoveryCode = &InputLoginFlowStepAuthenticate{}
var _ inputTakeAuthenticationCandidateIndex = &InputLoginFlowStepAuthenticate{}

func (*InputLoginFlowStepAuthenticate) Input() {}

func (i *InputLoginFlowStepAuthenticate) GetAuthenticationMethod() config.WorkflowAuthenticationMethod {
	return i.Authentication
}

func (i *InputLoginFlowStepAuthenticate) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}

func (i *InputLoginFlowStepAuthenticate) GetPassword() string {
	return i.Password
}

func (i *InputLoginFlowStepAuthenticate) GetCode() string {
	return i.Code
}

func (i *InputLoginFlowStepAuthenticate) GetRecoveryCode() string {
	return i.RecoveryCode
}

func (i *InputLoginFlowStepAuthenticate) GetIndex() int {
	return i.Index
}
