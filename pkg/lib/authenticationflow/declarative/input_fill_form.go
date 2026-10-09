package declarative

import (
	"context"
	"encoding/json"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaFillForm struct {
	JSONPointer    jsonpointer.T
	FlowRootObject config.AuthenticationFlowObject
	Config         *config.AppConfig
	Form           *config.AuthenticationFlowForm
}

var _ authflow.InputSchema = &InputSchemaFillForm{}
var _ authflow.InputSchemaWithFlowActionType = &InputSchemaFillForm{}

func (s *InputSchemaFillForm) GetJSONPointer() jsonpointer.T {
	return s.JSONPointer
}

func (s *InputSchemaFillForm) GetFlowRootObject() config.AuthenticationFlowObject {
	return s.FlowRootObject
}

// FlowActionType is fill_form even when the pointer is an identify branch.
func (s *InputSchemaFillForm) FlowActionType() authflow.FlowActionType {
	return authflow.FlowActionTypeFromStepType(config.AuthenticationFlowStepTypeFillForm)
}

func (s *InputSchemaFillForm) SchemaBuilder() validation.SchemaBuilder {
	// additionalProperties: false stops an identify input from parsing as a form input.
	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		AdditionalPropertiesFalse().
		Required("fields")
	b.Properties().Property("fields", buildFormFieldsSchema(s.Config, s.Form))
	return b
}

func (s *InputSchemaFillForm) MakeInput(ctx context.Context, rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputFillForm
	err := s.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(ctx, rawMessage, &input)
	if err != nil {
		return nil, err
	}

	// Report every form error before the flow reacts, so the flow stays in the step.
	vctx := &validation.Context{}
	validateFormFields(vctx, s.Config, input.Fields)
	err = vctx.Error("invalid form")
	if err != nil {
		return nil, err
	}

	return &input, nil
}

type InputFillForm struct {
	Fields []FormFieldInput `json:"fields"`
}

var _ authflow.Input = &InputFillForm{}
var _ inputFillForm = &InputFillForm{}

func (i *InputFillForm) Input() {}

func (i *InputFillForm) GetFormFields() []FormFieldInput {
	return i.Fields
}
