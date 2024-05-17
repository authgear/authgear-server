package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeOAuthAuthorizationResponse struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakeOAuthAuthorizationResponse{}

func (i *InputSchemaTakeOAuthAuthorizationResponse) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeOAuthAuthorizationResponse) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeOAuthAuthorizationResponse) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.Type(validation.TypeObject)
	b.Required("query")
	b.Properties().Property("query", validation.SchemaBuilder{}.Type(validation.TypeString))
	return b
}

func (i *InputSchemaTakeOAuthAuthorizationResponse) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeOAuthAuthorizationResponse
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeOAuthAuthorizationResponse struct {
	Query string `json:"query,omitempty"`
}

var _ authflow.Input = &InputTakeOAuthAuthorizationResponse{}
var _ inputTakeOAuthAuthorizationResponse = &InputTakeOAuthAuthorizationResponse{}

func (*InputTakeOAuthAuthorizationResponse) Input() {}

func (i *InputTakeOAuthAuthorizationResponse) GetQuery() string {
	return i.Query
}
