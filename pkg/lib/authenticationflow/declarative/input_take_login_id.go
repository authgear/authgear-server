package declarative

import (
	"encoding/json"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
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

var _ authflow.InputSchema = &InputTakeLoginID{}
var _ authflow.Input = &InputTakeLoginID{}
var _ inputTakeLoginID = &InputTakeLoginID{}

func (*InputTakeLoginID) SchemaBuilder() validation.SchemaBuilder {
	return InputTakeLoginIDSchemaBuilder
}

func (i *InputTakeLoginID) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
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
