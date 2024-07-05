package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaSetupTOTP struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaSetupTOTP{}

func (i *InputSchemaSetupTOTP) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaSetupTOTP) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaSetupTOTP) SchemaBuilder() validation.SchemaBuilder {
	inputSetupTOTPSchemaBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("code")

	inputSetupTOTPSchemaBuilder.Properties().Property(
		"code",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)

	return inputSetupTOTPSchemaBuilder
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
	Code string `json:"code,omitempty"`
}

var _ authflow.Input = &InputSetupTOTP{}
var _ inputSetupTOTP = &InputSetupTOTP{}

func (*InputSetupTOTP) Input() {}

func (i *InputSetupTOTP) GetCode() string {
	return i.Code
}
