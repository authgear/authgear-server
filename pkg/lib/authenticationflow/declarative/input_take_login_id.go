package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeLoginID struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
	BotProtectionCfg        *config.BotProtectionConfig
	IsExternalJWTAllowed    bool
}

var _ authflow.InputSchema = &InputSchemaTakeLoginID{}

func (i *InputSchemaTakeLoginID) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeLoginID) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeLoginID) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	if i.IsExternalJWTAllowed {
		b.OneOf(
			validation.SchemaBuilder{}.Required("login_id"),
			validation.SchemaBuilder{}.Required("external_jwt"),
		)
	} else {
		b.Required("login_id")
	}

	b.Properties().
		Property("login_id", validation.SchemaBuilder{}.Type(validation.TypeString))

	if i.IsExternalJWTAllowed {
		b.Properties().
			Property("external_jwt", validation.SchemaBuilder{}.Type(validation.TypeString))
	}

	if i.IsBotProtectionRequired && i.BotProtectionCfg != nil {
		b = AddBotProtectionToExistingSchemaBuilder(b, i.BotProtectionCfg)
	}
	return b
}

func (i *InputSchemaTakeLoginID) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeLoginID
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeLoginID struct {
	LoginID       string                      `json:"login_id,omitempty"`
	ExternalJWT   string                      `json:"external_jwt,omitempty"`
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeLoginID{}
var _ inputTakeLoginID = &InputTakeLoginID{}
var _ inputTakeLoginIDOrExternalJWT = &InputTakeLoginID{}
var _ inputTakeBotProtection = &InputTakeLoginID{}

func (*InputTakeLoginID) Input() {}

func (i *InputTakeLoginID) GetLoginID() string {
	return i.LoginID
}

func (i *InputTakeLoginID) GetExternalJWT() string {
	return i.ExternalJWT
}

func (i *InputTakeLoginID) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputTakeLoginID) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeLoginID) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
