package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaStepAccountRecoverySelectDestination struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	Options        []AccountRecoveryDestinationOption
}

var _ authflow.InputSchema = &InputSchemaStepAccountRecoverySelectDestination{}

func (i *InputSchemaStepAccountRecoverySelectDestination) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaStepAccountRecoverySelectDestination) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaStepAccountRecoverySelectDestination) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}
	indices := []interface{}{}
	for idx := range i.Options {
		indices = append(indices, idx)
	}
	b.Properties().Property("index", validation.SchemaBuilder{}.Type(validation.TypeInteger).Enum(indices...))
	b.Required("index")

	return b
}

func (i *InputSchemaStepAccountRecoverySelectDestination) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputStepAccountRecoverySelectDestination
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputStepAccountRecoverySelectDestination struct {
	Index int `json:"index,omitempty"`
}

var _ authflow.Input = &InputStepAccountRecoverySelectDestination{}
var _ inputTakeAccountRecoveryDestinationOptionIndex = &InputStepAccountRecoverySelectDestination{}

func (*InputStepAccountRecoverySelectDestination) Input() {}

func (i *InputStepAccountRecoverySelectDestination) GetAccountRecoveryDestinationOptionIndex() int {
	return i.Index
}
