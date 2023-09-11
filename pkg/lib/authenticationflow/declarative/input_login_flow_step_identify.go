package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaLoginFlowStepIdentify struct {
	JSONPointer jsonpointer.T
	OneOf       []*config.AuthenticationFlowLoginFlowOneOf
}

var _ authflow.InputSchema = &InputSchemaLoginFlowStepIdentify{}

func (i *InputSchemaLoginFlowStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaLoginFlowStepIdentify) SchemaBuilder() validation.SchemaBuilder {
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
		case config.AuthenticationFlowIdentificationOAuth:
			// FIXME(authflow): support oauth in login.
			continue
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

func (i *InputSchemaLoginFlowStepIdentify) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputLoginFlowStepIdentify
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputLoginFlowStepIdentify struct {
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`

	LoginID string `json:"login,omitempty"`
}

var _ authflow.Input = &InputLoginFlowStepIdentify{}
var _ inputTakeIdentificationMethod = &InputLoginFlowStepIdentify{}
var _ inputTakeLoginID = &InputLoginFlowStepIdentify{}

func (*InputLoginFlowStepIdentify) Input() {}

func (i *InputLoginFlowStepIdentify) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *InputLoginFlowStepIdentify) GetLoginID() string {
	return i.LoginID
}
