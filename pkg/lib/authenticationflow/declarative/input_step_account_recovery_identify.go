package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaStepAccountRecoveryIdentify struct {
	JSONPointer               jsonpointer.T
	FlowRootObject            config.AuthenticationFlowObject
	Options                   []AccountRecoveryIdentificationOption
	ShouldBypassBotProtection bool
}

var _ authflow.InputSchema = &InputSchemaStepAccountRecoveryIdentify{}

func (i *InputSchemaStepAccountRecoveryIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaStepAccountRecoveryIdentify) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaStepAccountRecoveryIdentify) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, option := range i.Options {
		b := validation.SchemaBuilder{}
		required := []string{"identification"}
		b.Properties().Property("identification", validation.SchemaBuilder{}.Const(option.Identification))

		requireString := func(key string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString))
		}
		requireBotProtection := func() {
			required = append(required, "bot_protection")
			b.Properties().Property("bot_protection", InputTakeBotProtectionBodySchemaBuilder)
		}

		setRequiredAndAppendOneOf := func() {
			b.Required(required...)
			oneOf = append(oneOf, b)
		}

		if !i.ShouldBypassBotProtection && option.isBotProtectionRequired() {
			requireBotProtection()
		}

		switch option.Identification {
		case config.AuthenticationFlowAccountRecoveryIdentificationEmail:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowAccountRecoveryIdentificationPhone:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		default:
			break
		}
	}

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	if len(oneOf) > 0 {
		b.OneOf(oneOf...)
	}

	return b
}

func (i *InputSchemaStepAccountRecoveryIdentify) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputStepAccountRecoveryIdentify
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputStepAccountRecoveryIdentify struct {
	Identification config.AuthenticationFlowAccountRecoveryIdentification `json:"identification,omitempty"`

	LoginID string `json:"login,omitempty"`
}

var _ authflow.Input = &InputStepIdentify{}
var _ inputTakeAccountRecoveryIdentificationMethod = &InputStepAccountRecoveryIdentify{}
var _ inputTakeLoginID = &InputStepAccountRecoveryIdentify{}

func (*InputStepAccountRecoveryIdentify) Input() {}

func (i *InputStepAccountRecoveryIdentify) GetAccountRecoveryIdentificationMethod() config.AuthenticationFlowAccountRecoveryIdentification {
	return i.Identification
}

func (i *InputStepAccountRecoveryIdentify) GetLoginID() string {
	return i.LoginID
}
