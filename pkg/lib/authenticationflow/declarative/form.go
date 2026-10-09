package declarative

import (
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

// addressSubKeys are the sub-attributes of /address.
var addressSubKeys = []string{"formatted", "street_address", "locality", "region", "postal_code", "country"}

// FormFieldInput is one item of the `fields` input. It is persisted in flow state.
type FormFieldInput struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func formFieldKey(f *config.AuthenticationFlowFormField) string {
	if f.UserProfile != nil {
		return f.UserProfile.Pointer
	}
	panic(fmt.Errorf("form field has no kind"))
}

func formFieldRequired(f *config.AuthenticationFlowFormField) bool {
	if f.UserProfile != nil {
		return f.UserProfile.Required
	}
	panic(fmt.Errorf("form field has no kind"))
}

// formFromFlowObject panics if o is not an AuthenticationFlowObjectFormProvider with a non-nil form.
func formFromFlowObject(o config.AuthenticationFlowObject) *config.AuthenticationFlowForm {
	provider, ok := o.(config.AuthenticationFlowObjectFormProvider)
	if !ok {
		panic(fmt.Errorf("flow object is %T, which cannot have a form", o))
	}
	form := provider.GetForm()
	if form == nil {
		panic(fmt.Errorf("flow object %T has no form", o))
	}
	return form
}

// findCustomAttribute returns nil if the pointer is not a custom attribute.
func findCustomAttribute(cfg *config.AppConfig, pointer string) *config.CustomAttributesAttributeConfig {
	for _, c := range cfg.UserProfile.CustomAttributes.Attributes {
		if c.Pointer == pointer {
			return c
		}
	}
	return nil
}

// formFieldValueSchema is the schema of the value of a field.
func formFieldValueSchema(cfg *config.AppConfig, pointer string) validation.SchemaBuilder {
	if b, ok := stdattrs.SchemaBuilderForPointerString(pointer); ok {
		switch pointer {
		case "/locale":
			b = b.Enum(slice.Cast[string, any](cfg.Localization.SupportedLanguages)...)
		case "/address":
			// The address attribute schema accepts unknown keys, which would be stored unvalidated.
			b = b.AdditionalPropertiesFalse()
		}
		return b
	}
	if c := findCustomAttribute(cfg, pointer); c != nil {
		b, err := c.ToSchemaBuilder()
		if err != nil {
			panic(fmt.Errorf("failed to construct custom attribute schema builder: %w", err))
		}
		return b
	}
	// Config validation has ensured the pointer is fillable.
	panic(fmt.Errorf("unknown attribute: %v", pointer))
}

// formFieldFilledSchema is the schema of a value that fills a required field.
func formFieldFilledSchema(cfg *config.AppConfig, pointer string, valueSchema validation.SchemaBuilder) validation.SchemaBuilder {
	if pointer == "/address" {
		// The address schema accepts an empty object.
		var anyOf []validation.SchemaBuilder
		for _, sub := range addressSubKeys {
			anyOf = append(anyOf, validation.SchemaBuilder{}.Required(sub))
		}
		return valueSchema.Clone().AnyOf(anyOf...)
	}
	if c := findCustomAttribute(cfg, pointer); c != nil && c.Type == config.CustomAttributeTypeBoolean {
		return validation.SchemaBuilder{}.Const(true)
	}
	// Strings have minLength 1, and numbers cannot be empty.
	return valueSchema
}

// buildFormFieldsSchema builds the schema of the `fields` array for form.
func buildFormFieldsSchema(cfg *config.AppConfig, form *config.AuthenticationFlowForm) validation.SchemaBuilder {
	keys := []any{}
	var itemsAllOf []validation.SchemaBuilder
	var fieldsAllOf []validation.SchemaBuilder

	for _, field := range form.Fields {
		key := formFieldKey(field)
		keys = append(keys, key)

		valueSchema := formFieldValueSchema(cfg, key)

		ifKey := validation.SchemaBuilder{}
		ifKey.Properties().Property("key", validation.SchemaBuilder{}.Const(key))
		thenValue := validation.SchemaBuilder{}
		thenValue.Properties().Property("value", valueSchema)
		itemsAllOf = append(itemsAllOf, validation.SchemaBuilder{}.If(ifKey).Then(thenValue))

		if formFieldRequired(field) {
			contains := validation.SchemaBuilder{}
			props := contains.Properties()
			props.Property("key", validation.SchemaBuilder{}.Const(key))
			props.Property("value", formFieldFilledSchema(cfg, key, valueSchema))
			fieldsAllOf = append(fieldsAllOf, validation.SchemaBuilder{}.Contains(contains))
		}
	}

	items := validation.SchemaBuilder{}.
		Type(validation.TypeObject).
		AdditionalPropertiesFalse().
		Required("key", "value")
	items.Properties().Property("key", validation.SchemaBuilder{}.Type(validation.TypeString).Enum(keys...))
	// The type of value is checked by itemsAllOf per key.
	items.Properties().Property("value", validation.SchemaBuilder{})
	if len(itemsAllOf) > 0 {
		items.AllOf(itemsAllOf...)
	}

	fields := validation.SchemaBuilder{}.
		Type(validation.TypeArray).
		Items(items)
	if len(fieldsAllOf) > 0 {
		fields.AllOf(fieldsAllOf...)
	}
	return fields
}

// validateFormFields checks what JSON Schema cannot: duplicate keys and string constraints.
// Every submitted value is checked, whether or not it changed.
func validateFormFields(vctx *validation.Context, cfg *config.AppConfig, fields []FormFieldInput) {
	seen := map[string]struct{}{}
	for i, f := range fields {
		index := strconv.Itoa(i)
		if _, ok := seen[f.Key]; ok {
			vctx.Child("fields", index, "key").EmitErrorMessage("duplicated key")
			continue
		}
		seen[f.Key] = struct{}{}

		valueCtx := vctx.Child("fields", index, "value")
		if f.Key == "/address" {
			address, ok := f.Value.(map[string]any)
			if !ok {
				continue
			}
			for _, sub := range addressSubKeys {
				validateStringValue(valueCtx.Child(sub), cfg.UserProfile.StandardAttributes.GetStringConstraints("/address/"+sub), "/address/"+sub, address[sub])
			}
			continue
		}

		var cons *config.StringConstraints
		if c := findCustomAttribute(cfg, f.Key); c != nil {
			cons = c.StringConstraints()
		} else {
			cons = cfg.UserProfile.StandardAttributes.GetStringConstraints(f.Key)
		}
		validateStringValue(valueCtx, cons, f.Key, f.Value)
	}
}

func validateStringValue(vctx *validation.Context, cons *config.StringConstraints, pointer string, value any) {
	s, ok := value.(string)
	if !ok || cons == nil {
		return
	}
	attrs.ValidateStringConstraints(vctx, cons, s, attrs.IsMultilineStandardAttribute(pointer))
}

// formFieldsToAttrs returns one attribute per form field, in form order.
// A field absent from the input gets a nil value, so its stored value is cleared.
func formFieldsToAttrs(cfg *config.AppConfig, form *config.AuthenticationFlowForm, fields []FormFieldInput) (std attrs.List, custom attrs.List) {
	submitted := map[string]any{}
	for _, f := range fields {
		submitted[f.Key] = f.Value
	}

	var all attrs.List
	for _, field := range form.Fields {
		key := formFieldKey(field)
		all = append(all, attrs.T{Pointer: key, Value: submitted[key]})
	}

	std, custom, unknown := all.Separate(cfg.UserProfile)
	if len(unknown) > 0 {
		panic(fmt.Errorf("config validation should have ensured there are no unknown attributes"))
	}
	return std, custom
}
