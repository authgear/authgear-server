package workflowconfig

import (
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaFillUserProfile struct {
	Attributes       []*config.WorkflowSignupFlowUserProfile
	CustomAttributes []*config.CustomAttributesAttributeConfig
}

var _ workflow.InputSchema = &InputSchemaFillUserProfile{}

func (s *InputSchemaFillUserProfile) buildCustomAttrPointerToItsConfigMap() map[string]*config.CustomAttributesAttributeConfig {
	customAttrPointerToItsConfig := make(map[string]*config.CustomAttributesAttributeConfig)
	for _, c := range s.CustomAttributes {
		c := c
		customAttrPointerToItsConfig[c.Pointer] = c
	}

	return customAttrPointerToItsConfig
}

func (s *InputSchemaFillUserProfile) buildContainsSchema(pointer string) validation.SchemaBuilder {
	contains := validation.SchemaBuilder{}
	contains.Properties().Property("pointer", validation.SchemaBuilder{}.Const(pointer))
	b := validation.SchemaBuilder{}.
		Contains(contains)
	return b
}

func (s *InputSchemaFillUserProfile) buildItemsSchema(pointer string, schema validation.SchemaBuilder) validation.SchemaBuilder {
	b := validation.SchemaBuilder{}
	if_ := validation.SchemaBuilder{}
	if_.Properties().Property("pointer", validation.SchemaBuilder{}.Const(pointer))
	then_ := validation.SchemaBuilder{}
	then_.Properties().Property("value", schema)
	b.If(if_).Then(then_)
	return b
}

func (s *InputSchemaFillUserProfile) SchemaBuilder() validation.SchemaBuilder {
	m := s.buildCustomAttrPointerToItsConfigMap()

	items := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("pointer", "value")

	var pointerEnum []string
	var itemsAllOf []validation.SchemaBuilder
	var allOfContains []validation.SchemaBuilder

	for _, attribute := range s.Attributes {
		pointerEnum = append(pointerEnum, attribute.Pointer)

		if stdAttrSchemaBuilder, ok := stdattrs.SchemaBuilderForPointerString(attribute.Pointer); ok {
			itemsAllOf = append(itemsAllOf, s.buildItemsSchema(attribute.Pointer, stdAttrSchemaBuilder))
		} else if cfg, ok := m[attribute.Pointer]; ok {
			customAttrSchemaBuilder, err := cfg.ToSchemaBuilder()
			if err != nil {
				panic(fmt.Errorf("failed to construct custom attribute schema builder: %w", err))
			}
			itemsAllOf = append(itemsAllOf, s.buildItemsSchema(attribute.Pointer, customAttrSchemaBuilder))
		} else {
			// Normally this branch should not be reachable.
		}

		if attribute.Required {
			allOfContains = append(allOfContains, s.buildContainsSchema(attribute.Pointer))
		}
	}

	items.Properties().Property("pointer", validation.SchemaBuilder{}.
		Type(validation.TypeString).
		Format("json-pointer").
		Enum(slice.Cast[string, interface{}](pointerEnum)...))
	items.AllOf(itemsAllOf...)

	b := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		Required("attributes")
	b.Properties().Property("attributes", validation.SchemaBuilder{}.
		Type(validation.TypeArray).
		Items(items).
		AllOf(allOfContains...),
	)
	return b
}

func (s *InputSchemaFillUserProfile) MakeInput(rawMessage json.RawMessage) (workflow.Input, error) {
	var input InputFillUserProfile
	err := s.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputFillUserProfile struct {
	Attributes []attrs.T `json:"attributes,omitempty"`
}

var _ workflow.Input = &InputFillUserProfile{}
var _ inputFillUserProfile = &InputFillUserProfile{}

func (i *InputFillUserProfile) Input() {}

func (i *InputFillUserProfile) GetAttributes() []attrs.T {
	return i.Attributes
}
