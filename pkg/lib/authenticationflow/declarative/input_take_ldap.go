package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeLDAP struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeLDAP{}

func (i *InputSchemaTakeLDAP) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeLDAP) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeLDAP) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Properties().
		Property(
			"server",
			validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1),
		).
		Property(
			"username",
			validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1),
		).
		Property(
			"password",
			validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1),
		).
		Required("server", "username", "password")
	return b
}

func (i *InputSchemaTakeLDAP) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeLDAP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeLDAP struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var _ authflow.Input = &InputTakeLDAP{}
var _ inputTakeLDAP = &InputTakeLDAP{}

func (*InputTakeLDAP) Input() {}

func (i *InputTakeLDAP) GetServer() string {
	return i.Server
}

func (i *InputTakeLDAP) GetUsername() string {
	return i.Username
}

func (i *InputTakeLDAP) GetPassword() string {
	return i.Password
}
