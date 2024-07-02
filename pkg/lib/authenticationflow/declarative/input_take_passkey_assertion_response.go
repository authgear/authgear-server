package declarative

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

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
}

type InputSchemaTakePasskeyAssertionResponse struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
}

var _ authflow.InputSchema = &InputSchemaTakePasskeyAssertionResponse{}

func (i *InputSchemaTakePasskeyAssertionResponse) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakePasskeyAssertionResponse) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakePasskeyAssertionResponse) SchemaBuilder() validation.SchemaBuilder {

	var inputSchemaTakePasskeyAssertionResponseSchemaBuilder = validation.SchemaBuilder{}.Type(validation.TypeObject)
	inputSchemaTakePasskeyAssertionResponseSchemaBuilder.Required("assertion_response")
	inputSchemaTakePasskeyAssertionResponseSchemaBuilder.Properties().Property("assertion_response", passkeyAssertionResponseSchemaBuilder)

	if i.IsBotProtectionRequired {
		inputSchemaTakePasskeyAssertionResponseSchemaBuilder = AddBotProtectionToExistingSchemaBuilder(inputSchemaTakePasskeyAssertionResponseSchemaBuilder)
	}

	return inputSchemaTakePasskeyAssertionResponseSchemaBuilder
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
	BotProtection     *InputTakeBotProtection               `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakePasskeyAssertionResponse{}
var _ inputTakePasskeyAssertionResponse = &InputTakePasskeyAssertionResponse{}
var _ inputTakeBotProtection = &InputTakePasskeyAssertionResponse{}

func (*InputTakePasskeyAssertionResponse) Input() {}

func (i *InputTakePasskeyAssertionResponse) GetAssertionResponse() *protocol.CredentialAssertionResponse {
	return i.AssertionResponse
}

func (i *InputTakePasskeyAssertionResponse) GetBotProtectionProvider() *InputTakeBotProtection {
	return i.BotProtection
}

func (i *InputTakePasskeyAssertionResponse) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakePasskeyAssertionResponse) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
