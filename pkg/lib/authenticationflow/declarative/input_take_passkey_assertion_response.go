package declarative

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var InputSchemaTakePasskeyAssertionResponseSchemaBuilder validation.SchemaBuilder
var passkeyAssertionResponseSchemaBuilder validation.SchemaBuilder

func init() {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	clientExtensionResults := validation.SchemaBuilder{}.Type(validation.TypeObject)

	base64URLString := validation.SchemaBuilder{}.Type(validation.TypeString).Format("x_base64_url")

	response := validation.SchemaBuilder{}.Type(validation.TypeObject)
	response.Properties().Property("clientDataJSON", base64URLString)
	response.Properties().Property("authenticatorData", base64URLString)
	response.Properties().Property("signature", base64URLString)
	// optional
	response.Properties().Property("userHandle", base64URLString)
	response.Required("clientDataJSON", "authenticatorData", "signature")

	b.Properties().Property("id", validation.SchemaBuilder{}.Type(validation.TypeString))
	b.Properties().Property("type", validation.SchemaBuilder{}.Type(validation.TypeString))
	b.Properties().Property("rawId", base64URLString)
	b.Properties().Property("clientExtensionResults", clientExtensionResults)
	b.Properties().Property("response", response)
	b.Required("id", "type", "rawId", "response")

	passkeyAssertionResponseSchemaBuilder = b

	InputSchemaTakePasskeyAssertionResponseSchemaBuilder = validation.SchemaBuilder{}.Type(validation.TypeObject)
	InputSchemaTakePasskeyAssertionResponseSchemaBuilder.Required("assertion_response")
	InputSchemaTakePasskeyAssertionResponseSchemaBuilder.Properties().Property("assertion_response", passkeyAssertionResponseSchemaBuilder)
}

type InputSchemaTakePasskeyAssertionResponse struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
}

var _ authflow.InputSchema = &InputSchemaTakePasskeyAssertionResponse{}

func (i *InputSchemaTakePasskeyAssertionResponse) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakePasskeyAssertionResponse) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (*InputSchemaTakePasskeyAssertionResponse) SchemaBuilder() validation.SchemaBuilder {
	return InputSchemaTakePasskeyAssertionResponseSchemaBuilder
}

func (i *InputSchemaTakePasskeyAssertionResponse) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakePasskeyAssertionResponse
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakePasskeyAssertionResponse struct {
	AssertionResponse *protocol.CredentialAssertionResponse `json:"assertion_response,omitempty"`
}

var _ authflow.Input = &InputTakePasskeyAssertionResponse{}
var _ inputTakePasskeyAssertionResponse = &InputTakePasskeyAssertionResponse{}

func (*InputTakePasskeyAssertionResponse) Input() {}

func (i *InputTakePasskeyAssertionResponse) GetAssertionResponse() *protocol.CredentialAssertionResponse {
	return i.AssertionResponse
}
