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

	good := validation.SchemaBuilder{}.Type(validation.TypeObject)
	good.Required("code")
	good.Properties().Property("code", validation.SchemaBuilder{}.Type(validation.TypeString))

	bad := validation.SchemaBuilder{}.Type(validation.TypeObject)
	bad.Required("error")
	good.Properties().Property("error", validation.SchemaBuilder{}.Type(validation.TypeString))
	good.Properties().Property("error_description", validation.SchemaBuilder{}.Type(validation.TypeString))
	good.Properties().Property("error_uri", validation.SchemaBuilder{}.Type(validation.TypeString).Format("uri"))

	b.OneOf(good, bad)

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
	Code             string `json:"code,omitempty"`
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

var _ authflow.Input = &InputTakeOAuthAuthorizationResponse{}
var _ inputTakeOAuthAuthorizationResponse = &InputTakeOAuthAuthorizationResponse{}

func (*InputTakeOAuthAuthorizationResponse) Input() {}

func (i *InputTakeOAuthAuthorizationResponse) GetOAuthAuthorizationCode() string {
	return i.Code
}

func (i *InputTakeOAuthAuthorizationResponse) GetOAuthError() string {
	return i.Error
}

func (i *InputTakeOAuthAuthorizationResponse) GetOAuthErrorDescription() string {
	return i.ErrorDescription
}

func (i *InputTakeOAuthAuthorizationResponse) GetOAuthErrorURI() string {
	return i.ErrorURI
}
