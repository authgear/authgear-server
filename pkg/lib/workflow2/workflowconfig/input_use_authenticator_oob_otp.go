package workflowconfig

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaUseAuthenticatorOOBOTP struct {
	Candidates []UseAuthenticationCandidate
}

var _ workflow.InputSchema = &InputSchemaUseAuthenticatorOOBOTP{}

func (i *InputSchemaUseAuthenticatorOOBOTP) SchemaBuilder() validation.SchemaBuilder {
	indice := []int{}
	for index, candidate := range i.Candidates {
		index := index
		candidate := candidate

		switch candidate.Authentication {
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
			fallthrough
		case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
			fallthrough
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
			fallthrough
		case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
			indice = append(indice, index)
		default:
			break
		}
	}

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	b.Properties().Property("index", validation.SchemaBuilder{}.
		Type(validation.TypeInteger).
		Enum(slice.Cast[int, interface{}](indice)...),
	)

	return b
}

func (i *InputSchemaUseAuthenticatorOOBOTP) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputUseAuthenticatorOOBOTP
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputUseAuthenticatorOOBOTP struct {
	Index int `json:"index"`
}

var _ workflow.Input = &InputUseAuthenticatorOOBOTP{}
var _ inputTakeAuthenticationCandidateIndex = &InputUseAuthenticatorOOBOTP{}

func (*InputUseAuthenticatorOOBOTP) Input() {}

func (i *InputUseAuthenticatorOOBOTP) GetIndex() int {
	return i.Index
}
