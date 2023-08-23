package declarative

import (
	"encoding/json"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaSignupFlowStepIdentify struct {
	OneOf []*config.AuthenticationFlowSignupFlowOneOf
}

var _ authflow.InputSchema = &InputSchemaSignupFlowStepIdentify{}

func (i *InputSchemaSignupFlowStepIdentify) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, branch := range i.OneOf {
		branch := branch

		b := validation.SchemaBuilder{}
		required := []string{"identification"}
		b.Properties().Property("identification", validation.SchemaBuilder{}.Const(branch.Identification))

		requireString := func(key string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString))
		}

		switch branch.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			requireString("login_id")
		case config.AuthenticationFlowIdentificationPhone:
			requireString("login_id")
		case config.AuthenticationFlowIdentificationUsername:
			requireString("login_id")
		default:
			// Skip the following code.
			continue
		}

		b.Required(required...)
		oneOf = append(oneOf, b)
	}

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	if len(oneOf) > 0 {
		b.OneOf(oneOf...)
	}

	return b
}

func (i *InputSchemaSignupFlowStepIdentify) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputSignupFlowStepIdentify
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputSignupFlowStepIdentify struct {
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`

	LoginID string `json:"login,omitempty"`
}

var _ authflow.Input = &InputSignupFlowStepIdentify{}
var _ inputTakeIdentificationMethod = &InputSignupFlowStepIdentify{}
var _ inputTakeLoginID = &InputSignupFlowStepIdentify{}

func (*InputSignupFlowStepIdentify) Input() {}

func (i *InputSignupFlowStepIdentify) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *InputSignupFlowStepIdentify) GetLoginID() string {
	return i.LoginID
}
