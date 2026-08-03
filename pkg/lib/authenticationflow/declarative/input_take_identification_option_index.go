package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeIdentificationOptionIndex struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
	BotProtectionCfg        *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaTakeIdentificationOptionIndex{}

func (i *InputSchemaTakeIdentificationOptionIndex) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeIdentificationOptionIndex) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeIdentificationOptionIndex) SchemaBuilder() validation.SchemaBuilder {
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("index")

	b.Properties().
		Property("index", validation.SchemaBuilder{}.Type(validation.TypeInteger))

	if i.IsBotProtectionRequired && i.BotProtectionCfg != nil {
		b = AddBotProtectionToExistingSchemaBuilder(b, i.BotProtectionCfg)
	}
	return b
}

func (i *InputSchemaTakeIdentificationOptionIndex) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeIdentificationOptionIndex
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeIdentificationOptionIndex struct {
	Index         int                         `json:"index,omitempty"`
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeIdentificationOptionIndex{}
var _ inputTakeIdentificationOptionIndex = &InputTakeIdentificationOptionIndex{}
var _ inputTakeBotProtection = &InputTakeIdentificationOptionIndex{}

func (*InputTakeIdentificationOptionIndex) Input() {}

func (i *InputTakeIdentificationOptionIndex) GetIdentificationOptionIndex() int {
	return i.Index
}

func (i *InputTakeIdentificationOptionIndex) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputTakeIdentificationOptionIndex) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeIdentificationOptionIndex) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
