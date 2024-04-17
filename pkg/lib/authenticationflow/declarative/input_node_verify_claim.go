package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaNodeVerifyClaim struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	OTPForm        otp.Form
}

var _ authflow.InputSchema = &InputSchemaNodeVerifyClaim{}

func (i *InputSchemaNodeVerifyClaim) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaNodeVerifyClaim) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaNodeVerifyClaim) SchemaBuilder() validation.SchemaBuilder {
	resend := validation.SchemaBuilder{}.
		Required("resend")
	resend.Properties().Property(
		"resend", validation.SchemaBuilder{}.Type(validation.TypeBoolean).Const(true),
	)

	oneOf := []validation.SchemaBuilder{resend}

	switch i.OTPForm {
	case otp.FormCode:
		code := validation.SchemaBuilder{}.
			Required("code")
		code.Properties().Property(
			"code", validation.SchemaBuilder{}.Type(validation.TypeString),
		).Property(
			"request_device_token", validation.SchemaBuilder{}.Type(validation.TypeBoolean),
		)

		oneOf = append(oneOf, code)
	case otp.FormLink:
		check := validation.SchemaBuilder{}.
			Required("check")
		check.Properties().Property(
			"check", validation.SchemaBuilder{}.Type(validation.TypeBoolean).Const(true),
		)

		oneOf = append(oneOf, check)
	}

	return validation.SchemaBuilder{}.Type(validation.TypeObject).OneOf(oneOf...)
}

func (i *InputSchemaNodeVerifyClaim) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputNodeVerifyClaim
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputNodeVerifyClaim struct {
	Code               string `json:"code,omitempty"`
	Resend             bool   `json:"resend,omitempty"`
	Check              bool   `json:"check,omitempty"`
	RequestDeviceToken bool   `json:"request_device_token,omitempty"`
}

var _ authflow.Input = &InputNodeVerifyClaim{}
var _ inputNodeVerifyClaim = &InputNodeVerifyClaim{}
var _ inputDeviceTokenRequested = &InputNodeVerifyClaim{}

func (*InputNodeVerifyClaim) Input() {}

func (i *InputNodeVerifyClaim) IsCode() bool {
	return i.Code != ""
}

func (i *InputNodeVerifyClaim) GetCode() string {
	return i.Code
}

func (i *InputNodeVerifyClaim) IsResend() bool {
	return i.Resend
}

func (i *InputNodeVerifyClaim) IsCheck() bool {
	return i.Check
}

func (i *InputNodeVerifyClaim) GetDeviceTokenRequested() bool {
	return i.RequestDeviceToken
}
