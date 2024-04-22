package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaLoginFlowStepAuthenticate struct {
	JSONPointer        jsonpointer.T
	FlowRootObject     config.AuthenticationFlowObject
	Options            []AuthenticateOption
	DeviceTokenEnabled bool
}

var _ authflow.InputSchema = &InputSchemaLoginFlowStepAuthenticate{}

func (i *InputSchemaLoginFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaLoginFlowStepAuthenticate) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaLoginFlowStepAuthenticate) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for index, option := range i.Options {
		index := index
		option := option

		b := validation.SchemaBuilder{}
		required := []string{"authentication"}
		b.Properties().Property("authentication", validation.SchemaBuilder{}.Const(option.Authentication))

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
		mayRequireChannel := func() {
			if len(option.Channels) > 1 {
				required = append(required, "channel")
				b.Properties().Property("channel", validation.SchemaBuilder{}.
					Type(validation.TypeString).
					Enum(slice.Cast[model.AuthenticatorOOBChannel, interface{}](option.Channels)...),
				)
			}
		}

		setRequiredAndAppendOneOf := func() {
			b.Required(required...)
			oneOf = append(oneOf, b)
		}

		switch option.Authentication {
		case config.AuthenticationFlowAuthenticationPrimaryPassword:
			requireString("password")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAuthenticationPrimaryPasskey:
			required = append(required, "assertion_response")
			b.Properties().Property("assertion_response", passkeyAssertionResponseSchemaBuilder)
			setRequiredAndAppendOneOf()

		case config.AuthenticationFlowAuthenticationSecondaryPassword:
			requireString("password")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAuthenticationSecondaryTOTP:
			requireString("code")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPEmail:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAuthenticationPrimaryOOBOTPSMS:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPEmail:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAuthenticationSecondaryOOBOTPSMS:
			requireIndex()
			mayRequireChannel()
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAuthenticationRecoveryCode:
			requireString("recovery_code")
			setRequiredAndAppendOneOf()
		default:
			break
		}
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

func (i *InputSchemaLoginFlowStepAuthenticate) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputLoginFlowStepAuthenticate
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputLoginFlowStepAuthenticate struct {
	Authentication     config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	RequestDeviceToken bool                                    `json:"request_device_token,omitempty"`
	Password           string                                  `json:"password,omitempty"`
	Code               string                                  `json:"code,omitempty"`
	RecoveryCode       string                                  `json:"recovery_code,omitempty"`
	Index              int                                     `json:"index,omitempty"`
	Channel            model.AuthenticatorOOBChannel           `json:"channel,omitempty"`
}

var _ authflow.Input = &InputLoginFlowStepAuthenticate{}
var _ inputTakeAuthenticationMethod = &InputLoginFlowStepAuthenticate{}
var _ inputDeviceTokenRequested = &InputLoginFlowStepAuthenticate{}
var _ inputTakePassword = &InputLoginFlowStepAuthenticate{}
var _ inputTakeTOTP = &InputLoginFlowStepAuthenticate{}
var _ inputTakeRecoveryCode = &InputLoginFlowStepAuthenticate{}
var _ inputTakeAuthenticationOptionIndex = &InputLoginFlowStepAuthenticate{}
var _ inputTakeOOBOTPChannel = &InputLoginFlowStepAuthenticate{}

func (*InputLoginFlowStepAuthenticate) Input() {}

func (i *InputLoginFlowStepAuthenticate) GetAuthenticationMethod() config.AuthenticationFlowAuthentication {
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

func (i *InputLoginFlowStepAuthenticate) GetChannel() model.AuthenticatorOOBChannel {
	return i.Channel
}
