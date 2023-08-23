package declarative

import (
	"encoding/json"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeOOBOTPTargetSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeOOBOTPTargetSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("target")

	InputTakeOOBOTPTargetSchemaBuilder.Properties().Property(
		"target",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
}

type InputTakeOOBOTPTarget struct {
	Target string `json:"target"`
}

var _ authflow.InputSchema = &InputTakeOOBOTPTarget{}
var _ authflow.Input = &InputTakeOOBOTPTarget{}
var _ inputTakeOOBOTPTarget = &InputTakeOOBOTPTarget{}

func (*InputTakeOOBOTPTarget) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeOOBOTPTargetSchemaBuilder
}

func (i *InputTakeOOBOTPTarget) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeOOBOTPTarget
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputTakeOOBOTPTarget) Input() {}

func (i *InputTakeOOBOTPTarget) GetTarget() string {
	return i.Target
}
