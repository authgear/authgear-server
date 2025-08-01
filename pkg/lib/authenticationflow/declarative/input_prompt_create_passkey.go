package declarative

import (
	"context"
	"encoding/json"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaPromptCreatePasskey struct {
	JSONPointer        jsonpointer.T
	FlowRootObject     config.AuthenticationFlowObject
	AllowDoNotAskAgain bool
}

var _ authflow.InputSchema = &InputSchemaPromptCreatePasskey{}

func (i *InputSchemaPromptCreatePasskey) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaPromptCreatePasskey) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaPromptCreatePasskey) SchemaBuilder() validation.SchemaBuilder {
	attestation := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	clientExtensionResults := validation.SchemaBuilder{}.Type(validation.TypeObject)

	base64URLString := validation.SchemaBuilder{}.Type(validation.TypeString).Format("x_base64_url")

	response := validation.SchemaBuilder{}.Type(validation.TypeObject)
	response.Properties().Property("attestationObject", base64URLString)
	response.Properties().Property("clientDataJSON", base64URLString)
	response.Required("attestationObject", "clientDataJSON")

	attestation.Properties().Property("id", validation.SchemaBuilder{}.Type(validation.TypeString))
	attestation.Properties().Property("type", validation.SchemaBuilder{}.Type(validation.TypeString))
	attestation.Properties().Property("rawId", base64URLString)
	attestation.Properties().Property("clientExtensionResults", clientExtensionResults)
	attestation.Properties().Property("response", response)
	attestation.Required("id", "type", "rawId", "response")

	oneOfAttestation := validation.SchemaBuilder{}.Type(validation.TypeObject)
	oneOfAttestation.Required("creation_response")
	oneOfAttestation.Properties().Property("creation_response", attestation)

	oneOfSkip := validation.SchemaBuilder{}.Type(validation.TypeObject)
	oneOfSkip.Required("skip")
	oneOfSkip.Properties().Property("skip", validation.SchemaBuilder{}.Type(validation.TypeBoolean))
	if i.AllowDoNotAskAgain {
		oneOfSkip.Properties().Property("do_not_ask_again", validation.SchemaBuilder{}.Type(validation.TypeBoolean))
	}
	root := validation.SchemaBuilder{}.Type(validation.TypeObject)
	root.OneOf(oneOfAttestation, oneOfSkip)
	return root
}

func (i *InputSchemaPromptCreatePasskey) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputPromptCreatePasskey
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputPromptCreatePasskey struct {
	Skip             bool                                 `json:"skip,omitempty"`
	DoNotAskAgain    bool                                 `json:"do_not_ask_again,omitempty"`
	CreationResponse *protocol.CredentialCreationResponse `json:"creation_response,omitempty"`
}

var _ authflow.Input = &InputPromptCreatePasskey{}
var _ inputNodePromptCreatePasskey = &InputPromptCreatePasskey{}

func (*InputPromptCreatePasskey) Input() {}

func (i *InputPromptCreatePasskey) IsSkip() bool {
	return i.Skip
}

func (i *InputPromptCreatePasskey) IsDoNotAskAgain() bool {
	return i.DoNotAskAgain
}

func (i *InputPromptCreatePasskey) IsCreationResponse() bool {
	return i.CreationResponse != nil
}

func (i *InputPromptCreatePasskey) GetCreationResponse() *protocol.CredentialCreationResponse {
	return i.CreationResponse
}
