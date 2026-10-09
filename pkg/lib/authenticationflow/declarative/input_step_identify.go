package declarative

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaStepIdentify struct {
	JSONPointer               jsonpointer.T
	FlowRootObject            config.AuthenticationFlowObject
	Options                   []IdentificationOption
	ShouldBypassBotProtection bool
	BotProtectionCfg          *config.BotProtectionConfig
	IsExternalJWTAllowed      bool

	// Forms is keyed by identification. It is non-nil only where a branch can have a form,
	// which is in signup and promote flows. Only email, phone and username can have forms.
	// Options cannot be used to look a form up by branch,
	// because OAuth expands to one option per provider, and LDAP to one per server.
	Forms  map[model.AuthenticationFlowIdentification]*config.AuthenticationFlowForm
	Config *config.AppConfig
}

var _ authflow.InputSchema = &InputSchemaStepIdentify{}

func (i *InputSchemaStepIdentify) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaStepIdentify) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaStepIdentify) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for index, option := range i.Options {
		b := validation.SchemaBuilder{}
		required := []string{"identification"}
		b.Properties().Property("identification", validation.SchemaBuilder{}.Const(option.Identification))

		requireString := func(key string) {
			required = append(required, key)
			b.Properties().Property(key, validation.SchemaBuilder{}.Type(validation.TypeString))
		}
		requireBotProtection := func() {
			required = append(required, "bot_protection")
			b.Properties().Property("bot_protection", NewBotProtectionBodySchemaBuilder(i.BotProtectionCfg))
		}

		requireOneOfString := func(key1 string, key2 string) {
			b.Properties().Property(key1, validation.SchemaBuilder{}.Type(validation.TypeString))
			b.Properties().Property(key2, validation.SchemaBuilder{}.Type(validation.TypeString))
			b.OneOf(
				validation.SchemaBuilder{}.Required(key1),
				validation.SchemaBuilder{}.Required(key2),
			)
		}

		setRequiredAndAppendOneOf := func() {
			if i.setForm(b, option.Identification) {
				required = append(required, "fields")
			}
			b.Required(required...)
			oneOf = append(oneOf, b)
		}

		if !i.ShouldBypassBotProtection && i.BotProtectionCfg != nil && option.isBotProtectionRequired() {
			requireBotProtection()
		}

		switch option.Identification {
		case model.AuthenticationFlowIdentificationIDToken:
			requireString("id_token")
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowIdentificationEmail:
			if i.IsExternalJWTAllowed {
				requireOneOfString("login_id", "external_jwt")
			} else {
				requireString("login_id")
			}
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowIdentificationPhone:
			if i.IsExternalJWTAllowed {
				requireOneOfString("login_id", "external_jwt")
			} else {
				requireString("login_id")
			}
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowIdentificationUsername:
			if i.IsExternalJWTAllowed {
				requireOneOfString("login_id", "external_jwt")
			} else {
				requireString("login_id")
			}
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowIdentificationOAuth:
			// redirect_uri is required.
			required = append(required, "redirect_uri")
			b.Properties().Property("redirect_uri", validation.SchemaBuilder{}.Type(validation.TypeString).Format("uri"))

			// alias is required.
			required = append(required, "alias")
			b.Properties().Property("alias", validation.SchemaBuilder{}.Type(validation.TypeString).Const(option.Alias))

			// response_mode is optional.
			b.Properties().Property("response_mode", validation.SchemaBuilder{}.
				Type(validation.TypeString).
				Enum(oauthrelyingparty.ResponseModeFormPost, oauthrelyingparty.ResponseModeQuery))

			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowIdentificationPasskey:
			required = append(required, "assertion_response")
			b.Properties().Property("assertion_response", passkeyAssertionResponseSchemaBuilder)
			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowIdentificationLDAP:
			required = append(required, "server_name")
			b.Properties().
				Property(
					"server_name",
					validation.SchemaBuilder{}.Type(validation.TypeString).Const(option.ServerName),
				)

			required = append(required, "username")
			b.Properties().
				Property(
					"username",
					validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1),
				)

			required = append(required, "password")
			b.Properties().
				Property(
					"password",
					validation.SchemaBuilder{}.Type(validation.TypeString).MinLength(1),
				)

			setRequiredAndAppendOneOf()
		case model.AuthenticationFlowIdentificationSelectAccount:
			required = append(required, "index")
			b.Properties().Property("index", validation.SchemaBuilder{}.Type(validation.TypeInteger).Const(index))
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

// setForm adds the fields of the form of the identification to the schema b of an option.
// It reports whether fields is required.
func (i *InputSchemaStepIdentify) setForm(b validation.SchemaBuilder, identification model.AuthenticationFlowIdentification) (required bool) {
	if i.Forms == nil {
		return false
	}
	form := i.Forms[identification]
	if form == nil {
		// Reject fields rather than silently ignore them.
		b.Not(validation.SchemaBuilder{}.Required("fields"))
		return false
	}
	b.Properties().Property("fields", buildFormFieldsSchema(i.Config, form))
	return slices.ContainsFunc(form.Fields, formFieldRequired)
}

func (i *InputSchemaStepIdentify) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputStepIdentify
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	// Inputs made by code, such as synthetic inputs, leave it false.
	input.IsFieldsSubmitted = true

	if i.Forms[input.Identification] != nil {
		vctx := &validation.Context{}
		validateFormFields(vctx, i.Config, input.Fields)
		err = vctx.Error("invalid form")
		if err != nil {
			return nil, err
		}
	}
	return &input, nil
}

type InputStepIdentify struct {
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`

	IDToken string `json:"id_token,omitempty"`

	LoginID     string `json:"login_id,omitempty"`
	ExternalJWT string `json:"external_jwt,omitempty"`

	Alias        string `json:"alias,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
	ResponseMode string `json:"response_mode,omitempty"`

	BotProtection *InputTakeBotProtectionBody `json:"bot_protection,omitempty"`

	ServerName string `json:"server_name"`
	Username   string `json:"username"`
	Password   string `json:"password"`

	Index int `json:"index,omitempty"`

	// Fields is specific to the email, phone and username options that have a form.
	Fields []FormFieldInput `json:"fields,omitempty"`
	// IsFieldsSubmitted is true when the input was parsed from a request, whether or not it has fields.
	IsFieldsSubmitted bool `json:"-"`
}

var _ authflow.Input = &InputStepIdentify{}
var _ inputTakeIdentificationMethod = &InputStepIdentify{}
var _ inputTakeIDToken = &InputStepIdentify{}
var _ inputTakeLoginID = &InputStepIdentify{}
var _ inputTakeLoginIDOrExternalJWT = &InputStepIdentify{}
var _ inputTakeOAuthAuthorizationRequest = &InputStepIdentify{}
var _ inputTakeBotProtection = &InputStepIdentify{}
var _ inputTakeLDAP = &InputStepIdentify{}
var _ inputTakeIdentificationOptionIndex = &InputStepIdentify{}
var _ inputIdentifyFormFields = &InputStepIdentify{}

func (*InputStepIdentify) Input() {}

func (i *InputStepIdentify) GetIdentificationMethod() model.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *InputStepIdentify) GetIDToken() string {
	return i.IDToken
}

func (i *InputStepIdentify) GetLoginID() string {
	return i.LoginID
}

func (i *InputStepIdentify) GetExternalJWT() string {
	return i.ExternalJWT
}

func (i *InputStepIdentify) GetOAuthAlias() string {
	return i.Alias
}

func (i *InputStepIdentify) GetOAuthRedirectURI() string {
	return i.RedirectURI
}

func (i *InputStepIdentify) GetOAuthResponseMode() string {
	return i.ResponseMode
}

func (i *InputStepIdentify) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *InputStepIdentify) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *InputStepIdentify) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}

func (i *InputStepIdentify) GetServerName() string {
	return i.ServerName
}

func (i *InputStepIdentify) GetUsername() string {
	return i.Username
}

func (i *InputStepIdentify) GetPassword() string {
	return i.Password
}

func (i *InputStepIdentify) GetIdentificationOptionIndex() int {
	return i.Index
}

func (i *InputStepIdentify) GetIdentifyFormFields() (fields []FormFieldInput, submitted bool) {
	return i.Fields, i.IsFieldsSubmitted
}
