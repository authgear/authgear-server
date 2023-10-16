package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputCheckAccountStatusSchemaBuilder validation.SchemaBuilder

func init() {
	InputCheckAccountStatusSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject)
}

type InputSchemaCheckAccountStatus struct {
	JSONPointer jsonpointer.T
}

var _ authflow.InputSchema = &InputSchemaCheckAccountStatus{}

func (i *InputSchemaCheckAccountStatus) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (*InputSchemaCheckAccountStatus) SchemaBuilder() validation.SchemaBuilder {
	return InputCheckAccountStatusSchemaBuilder
}

func (i *InputSchemaCheckAccountStatus) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputCheckAccountStatus
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputCheckAccountStatus struct{}

var _ authflow.Input = &InputCheckAccountStatus{}

func (*InputCheckAccountStatus) Input() {}
