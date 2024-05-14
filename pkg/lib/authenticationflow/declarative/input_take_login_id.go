package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputTakeLoginIDSchemaBuilder validation.SchemaBuilder

func init() {
	InputTakeLoginIDSchemaBuilder = validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("login_id")

	InputTakeLoginIDSchemaBuilder.Properties().Property(
		"login_id",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
}

type InputSchemaTakeLoginID struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeLoginID{}

func (i *InputSchemaTakeLoginID) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeLoginID) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakeLoginID) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeLoginIDSchemaBuilder
}

func (i *InputSchemaTakeLoginID) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeLoginID
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeLoginID struct {
	LoginID string `json:"login_id"`
}

var _ authflow.Input = &InputTakeLoginID{}
var _ inputTakeLoginID = &InputTakeLoginID{}

func (*InputTakeLoginID) Input() {}

func (i *InputTakeLoginID) GetLoginID() string {
	return i.LoginID
}
