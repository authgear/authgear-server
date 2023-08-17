package workflowconfig

import (
	"encoding/json"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputSetupTOTPSchemaBuilder validation.SchemaBuilder

func init() {
	InputSetupTOTPSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("code", "display_name")

	InputSetupTOTPSchemaBuilder.Properties().Property(
		"code",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	InputSetupTOTPSchemaBuilder.Properties().Property(
		"display_name",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
}

type InputSetupTOTP struct {
	Code        string `json:"code,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

var _ workflow.InputSchema = &InputSetupTOTP{}
var _ workflow.Input = &InputSetupTOTP{}
var _ inputSetupTOTP = &InputSetupTOTP{}

func (*InputSetupTOTP) SchemaBuilder() validation.SchemaBuilder {
	return InputSetupTOTPSchemaBuilder
}

func (i *InputSetupTOTP) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputSetupTOTP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputSetupTOTP) Input() {}

func (i *InputSetupTOTP) GetCode() string {
	return i.Code
}

func (i *InputSetupTOTP) GetDisplayName() string {
	return i.DisplayName
}
