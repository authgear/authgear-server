package workflowconfig

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaLoginFlowStepIdentify struct {
	OneOf []*config.WorkflowLoginFlowOneOf
}

var _ workflow.InputSchema = &InputSchemaLoginFlowStepIdentify{}

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
		case config.WorkflowIdentificationMethodEmail:
			requireString("login_id")
		case config.WorkflowIdentificationMethodPhone:
			requireString("login_id")
		case config.WorkflowIdentificationMethodUsername:
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

func (i *InputSchemaLoginFlowStepIdentify) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputLoginFlowStepIdentify
	err := i.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputLoginFlowStepIdentify struct {
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`

	LoginID string `json:"login,omitempty"`
}

var _ workflow.Input = &InputLoginFlowStepIdentify{}
var _ inputTakeIdentificationMethod = &InputLoginFlowStepIdentify{}
var _ inputTakeLoginID = &InputLoginFlowStepIdentify{}

func (*InputLoginFlowStepIdentify) Input() {}

func (i *InputLoginFlowStepIdentify) GetIdentificationMethod() config.WorkflowIdentificationMethod {
	return i.Identification
}

func (i *InputLoginFlowStepIdentify) GetLoginID() string {
	return i.LoginID
}
