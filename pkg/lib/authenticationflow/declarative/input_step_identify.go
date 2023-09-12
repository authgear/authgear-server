package declarative

import (
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaStepIdentify struct {
	JSONPointer jsonpointer.T
	Candidates  []IdentificationCandidate
}

var _ authflow.InputSchema = &InputSchemaStepIdentify{}

func (i *InputSchemaStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaStepIdentify) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for _, candidate := range i.Candidates {
		b := validation.SchemaBuilder{}
		required := []string{"identification"}
		b.Properties().Property("identification", validation.SchemaBuilder{}.Const(candidate.Identification))

		requireString := func(key string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString))
		}

		requireFormat := func(key string, format string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString).Format(format))
		}

		requireConst := func(key string, value string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString).Const(value))
		}

		setRequiredAndAppendOneOf := func() {
			b.Required(required...)
			oneOf = append(oneOf, b)
		}

		switch candidate.Identification {
		case config.AuthenticationFlowIdentificationEmail:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationPhone:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationUsername:
			requireString("login_id")
			setRequiredAndAppendOneOf()
		case config.AuthenticationFlowIdentificationOAuth:
			requireFormat("redirect_uri", "uri")
			requireString("state")
			requireConst("alias", candidate.Alias)
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

func (i *InputSchemaStepIdentify) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputStepIdentify
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputStepIdentify struct {
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`

	LoginID string `json:"login,omitempty"`

	Alias       string `json:"alias,omitempty"`
	State       string `json:"state,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

var _ authflow.Input = &InputStepIdentify{}
var _ inputTakeIdentificationMethod = &InputStepIdentify{}
var _ inputTakeLoginID = &InputStepIdentify{}
var _ inputTakeOAuthAuthorizationRequest = &InputStepIdentify{}

func (*InputStepIdentify) Input() {}

func (i *InputStepIdentify) GetIdentificationMethod() config.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *InputStepIdentify) GetLoginID() string {
	return i.LoginID
}

func (i *InputStepIdentify) GetOAuthAlias() string {
	return i.Alias
}

func (i *InputStepIdentify) GetOAuthState() string {
	return i.State
}

func (i *InputStepIdentify) GetOAuthRedirectURI() string {
	return i.RedirectURI
}
