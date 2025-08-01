package declarative

import (
	"context"
	"encoding/json"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaAccountLinkingIdentification struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	Options        []AccountLinkingIdentificationOptionInternal
}

var _ authflow.InputSchema = &InputSchemaAccountLinkingIdentification{}

func (i *InputSchemaAccountLinkingIdentification) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

func (i *InputSchemaAccountLinkingIdentification) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (i *InputSchemaAccountLinkingIdentification) SchemaBuilder() validation.SchemaBuilder {
	oneOf := []validation.SchemaBuilder{}

	for index, option := range i.Options {
		index := index
		option := option
		b := validation.SchemaBuilder{}
		required := []string{"index"}
		b.Properties().Property("index", validation.SchemaBuilder{}.Const(index))
		switch option.Identifcation {
		case model.AuthenticationFlowIdentificationOAuth:
			required = append(required, "redirect_uri")
			b.Properties().Property("redirect_uri", validation.SchemaBuilder{}.Type(validation.TypeString).Format("uri"))
			// response_mode is optional.
			b.Properties().Property("response_mode", validation.SchemaBuilder{}.
				Type(validation.TypeString).
				Enum(oauthrelyingparty.ResponseModeFormPost, oauthrelyingparty.ResponseModeQuery))
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

func (i *InputSchemaAccountLinkingIdentification) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputAccountLinkingIdentification
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputAccountLinkingIdentification struct {
	Index int `json:"index,omitempty"`

	RedirectURI  string `json:"redirect_uri,omitempty"`
	ResponseMode string `json:"response_mode,omitempty"`
}

var _ authflow.Input = &InputAccountLinkingIdentification{}
var _ inputTakeAccountLinkingIdentification = &InputAccountLinkingIdentification{}

func (*InputAccountLinkingIdentification) Input() {}

func (i *InputAccountLinkingIdentification) GetAccountLinkingIdentificationIndex() int {
	return i.Index
}
func (i *InputAccountLinkingIdentification) GetAccountLinkingOAuthRedirectURI() string {
	return i.RedirectURI
}
func (i *InputAccountLinkingIdentification) GetAccountLinkingOAuthResponseMode() string {
	return i.ResponseMode
}
