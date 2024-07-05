package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
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

type InputSchemaTakeOOBOTPTarget struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeOOBOTPTarget{}

func (i *InputSchemaTakeOOBOTPTarget) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeOOBOTPTarget) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakeOOBOTPTarget) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeOOBOTPTargetSchemaBuilder
}

func (i *InputSchemaTakeOOBOTPTarget) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeOOBOTPTarget
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeOOBOTPTarget struct {
	Target string `json:"target"`
}

var _ authflow.Input = &InputTakeOOBOTPTarget{}
var _ inputTakeOOBOTPTarget = &InputTakeOOBOTPTarget{}

func (*InputTakeOOBOTPTarget) Input() {}

func (i *InputTakeOOBOTPTarget) GetTarget() string {
	return i.Target
}
