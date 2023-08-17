package workflowconfig

import (
	"encoding/json"

	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
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

type InputTakeLoginID struct {
	LoginID string `json:"login_id"`
}

var _ workflow.InputSchema = &InputTakeLoginID{}
var _ workflow.Input = &InputTakeLoginID{}
var _ inputTakeLoginID = &InputTakeLoginID{}

func (*InputTakeLoginID) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeLoginIDSchemaBuilder
}

func (i *InputTakeLoginID) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputTakeLoginID
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

func (*InputTakeLoginID) Input() {}

func (i *InputTakeLoginID) GetLoginID() string {
	return i.LoginID
}
