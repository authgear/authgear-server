package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
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

type InputSchemaSetupTOTP struct {
	JSONPointer jsonpointer.T
}

var _ authflow.InputSchema = &InputSchemaSetupTOTP{}

func (i *InputSchemaSetupTOTP) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (*InputSchemaSetupTOTP) SchemaBuilder() validation.SchemaBuilder {
	return InputSetupTOTPSchemaBuilder
}

func (i *InputSchemaSetupTOTP) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputSetupTOTP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputSetupTOTP struct {
	Code        string `json:"code,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

var _ authflow.Input = &InputSetupTOTP{}
var _ inputSetupTOTP = &InputSetupTOTP{}

func (*InputSetupTOTP) Input() {}

func (i *InputSetupTOTP) GetCode() string {
	return i.Code
}

func (i *InputSetupTOTP) GetDisplayName() string {
	return i.DisplayName
}
