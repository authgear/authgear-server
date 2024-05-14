package declarative

import (
	"encoding/json"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type InputSchemaFillInUserProfile struct {
	JSONPointer      jsonpointer.T
	FlowRootObject   config.AuthenticationFlowObject
	Attributes       []*config.AuthenticationFlowSignupFlowUserProfile
	CustomAttributes []*config.CustomAttributesAttributeConfig
}

var _ authflow.InputSchema = &InputSchemaFillInUserProfile{}

func (s *InputSchemaFillInUserProfile) buildCustomAttrPointerToItsConfigMap() map[string]*config.CustomAttributesAttributeConfig {
	customAttrPointerToItsConfig := make(map[string]*config.CustomAttributesAttributeConfig)
	for _, c := range s.CustomAttributes {
		c := c
		customAttrPointerToItsConfig[c.Pointer] = c
	}

	return customAttrPointerToItsConfig
}

func (s *InputSchemaFillInUserProfile) buildContainsSchema(pointer string) validation.SchemaBuilder {
	contains := validation.SchemaBuilder{}
	contains.Properties().Property("pointer", validation.SchemaBuilder{}.Const(pointer))
	b := validation.SchemaBuilder{}.
		Contains(contains)
	return b
}

func (s *InputSchemaFillInUserProfile) buildItemsSchema(pointer string, schema validation.SchemaBuilder) validation.SchemaBuilder {
	b := validation.SchemaBuilder{}
	if_ := validation.SchemaBuilder{}
	if_.Properties().Property("pointer", validation.SchemaBuilder{}.Const(pointer))
	then_ := validation.SchemaBuilder{}
	then_.Properties().Property("value", schema)
	b.If(if_).Then(then_)
	return b
}

func (s *InputSchemaFillInUserProfile) GetJSONPointer() jsonpointer.T {
	return s.JSONPointer
}

func (i *InputSchemaFillInUserProfile) GetFlowRootObject() config.AuthenticationFlowObject {
	return i.FlowRootObject
}

func (s *InputSchemaFillInUserProfile) SchemaBuilder() validation.SchemaBuilder {
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

func (s *InputSchemaFillInUserProfile) MakeInput(rawMessage json.RawMessage) (authflow.Input, error) {
	var input InputFillInUserProfile
	err := s.SchemaBuilder().ToSimpleSchema().Validator().ParseJSONRawMessage(rawMessage, &input)
	if err != nil {
		return nil, err
	}
	return &input, nil
}

type InputFillInUserProfile struct {
	Attributes []attrs.T `json:"attributes,omitempty"`
}

var _ authflow.Input = &InputFillInUserProfile{}
var _ inputFillInUserProfile = &InputFillInUserProfile{}

func (i *InputFillInUserProfile) Input() {}

func (i *InputFillInUserProfile) GetAttributes() []attrs.T {
	return i.Attributes
}
