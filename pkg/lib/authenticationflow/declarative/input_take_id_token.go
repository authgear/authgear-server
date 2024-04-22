package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeIDTokenSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeIDTokenSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("id_token")

	InputTakeIDTokenSchemaBuilder.Properties().Property(
		"id_token",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
}

type InputSchemaTakeIDToken struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeIDToken{}

func (i *InputSchemaTakeIDToken) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeIDToken) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakeIDToken) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeIDTokenSchemaBuilder
}

func (i *InputSchemaTakeIDToken) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeIDToken
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeIDToken struct {
	IDToken string `json:"id_token"`
}

var _ authflow.Input = &InputTakeIDToken{}
var _ inputTakeIDToken = &InputTakeIDToken{}

func (*InputTakeIDToken) Input() {}

func (i *InputTakeIDToken) GetIDToken() string {
	return i.IDToken
}
