package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaStepAccountRecoverySelectDestination struct {
	JSONPointer jsonpointer.T
	Options     []AccountRecoveryDestinationOption
}

var _ authflow.InputSchema = &InputSchemaStepAccountRecoverySelectDestination{}

func (i *InputSchemaStepAccountRecoverySelectDestination) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaStepAccountRecoverySelectDestination) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}
	ids := []interface{}{}
	for _, op := range i.Options {
		ids = append(ids, op.ID)
	}
	b.Properties().Property("option_id", validation.SchemaBuilder{}.Type(validation.TypeString).Enum(ids...))
	b.Required("option_id")

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
	OptionID string `json:"option_id,omitempty"`
}

var _ authflow.Input = &InputStepAccountRecoverySelectDestination{}
var _ inputTakeAccountRecoveryDestinationOptionID = &InputStepAccountRecoverySelectDestination{}

func (*InputStepAccountRecoverySelectDestination) Input() {}

func (i *InputStepAccountRecoverySelectDestination) GetAccountRecoveryDestinationOptionID() string {
	return i.OptionID
}
