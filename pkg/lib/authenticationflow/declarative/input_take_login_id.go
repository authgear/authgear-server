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
}

var _ authflow.InputSchema = &InputSchemaTakeLoginID{}

func (i *InputSchemaTakeLoginID) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeLoginID) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeLoginID) SchemaBuilder() validation.SchemaBuilder {
	inputTakeLoginIDSchemaBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("login_id")

	inputTakeLoginIDSchemaBuilder.Properties().Property(
		"login_id",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)
	if i.IsBotProtectionRequired && i.BotProtectionCfg != nil {
		inputTakeLoginIDSchemaBuilder = AddBotProtectionToExistingSchemaBuilder(inputTakeLoginIDSchemaBuilder, i.BotProtectionCfg)
	}
	return inputTakeLoginIDSchemaBuilder
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
	LoginID       string                      `json:"login_id"`
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeLoginID{}
var _ inputTakeLoginID = &InputTakeLoginID{}
var _ inputTakeBotProtection = &InputTakeLoginID{}

func (*InputTakeLoginID) Input() {}

func (i *InputTakeLoginID) GetLoginID() string {
	return i.LoginID
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
