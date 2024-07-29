package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaTakeOOBOTPTarget struct {
	JSONPointer             jsonpointer.T
	FlowRootObject          config.AuthenticationFlowObject
	IsBotProtectionRequired bool
	BotProtectionCfg        *config.BotProtectionConfig
}

var _ authflow.InputSchema = &InputSchemaTakeOOBOTPTarget{}

func (i *InputSchemaTakeOOBOTPTarget) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaTakeOOBOTPTarget) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaTakeOOBOTPTarget) SchemaBuilder() validation.SchemaBuilder {
	sb := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("target")

	sb.Properties().Property(
		"target",
		validation.SchemaBuilder{}.Type(validation.TypeString),
	)

	if i.IsBotProtectionRequired && i.BotProtectionCfg != nil {
		sb = AddBotProtectionToExistingSchemaBuilder(sb, i.BotProtectionCfg)
	}
	return sb
}

func (i *InputSchemaTakeOOBOTPTarget) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputTakeOOBOTPTarget
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputTakeOOBOTPTarget struct {
	Target        string                      `json:"target"`
	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &InputTakeOOBOTPTarget{}
var _ inputTakeOOBOTPTarget = &InputTakeOOBOTPTarget{}
var _ inputTakeBotProtection = &InputTakeOOBOTPTarget{}

func (*InputTakeOOBOTPTarget) Input() {}

func (i *InputTakeOOBOTPTarget) GetTarget() string {
	return i.Target
}

func (i *InputTakeOOBOTPTarget) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputTakeOOBOTPTarget) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputTakeOOBOTPTarget) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
