package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/htmlutil"
	"github.com/authgear/authgear-server/pkg/util/labelutil"
)

type FillFormFieldType string

const (
	FillFormFieldTypeString      FillFormFieldType = "string"
	FillFormFieldTypeInteger     FillFormFieldType = "integer"
	FillFormFieldTypeNumber      FillFormFieldType = "number"
	FillFormFieldTypeBoolean     FillFormFieldType = "boolean"
	FillFormFieldTypeEnum        FillFormFieldType = "enum"
	FillFormFieldTypeDate        FillFormFieldType = "date"
	FillFormFieldTypeEmail       FillFormFieldType = "email"
	FillFormFieldTypePhoneNumber FillFormFieldType = "phone_number"
	FillFormFieldTypeURL         FillFormFieldType = "url"
	FillFormFieldTypeCountryCode FillFormFieldType = "country_code"
	FillFormFieldTypeTimezone    FillFormFieldType = "timezone"
	FillFormFieldTypeAddress     FillFormFieldType = "address"
)

type FillFormData struct {
	TypedData
	Form FillFormDataForm `json:"form"`
}

type FillFormDataForm struct {
	Name   string              `json:"name,omitempty"`
	Fields []FillFormDataField `json:"fields"`
}

type FillFormDataField struct {
	Key      string            `json:"key"`
	Type     FillFormFieldType `json:"type"`
	Label    string            `json:"label"`
	Required bool              `json:"required"`
	// Value is omitted only when nil. A false is kept.
	Value       any                           `json:"value,omitempty"`
	Options     []FillFormDataFieldOption     `json:"options,omitempty"`
	Suggestions []string                      `json:"suggestions,omitempty"`
	Constraints *FillFormDataFieldConstraints `json:"constraints,omitempty"`
}

type FillFormDataFieldOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type FillFormDataFieldConstraints struct {
	MinLength         *int     `json:"min_length,omitempty"`
	MaxLength         *int     `json:"max_length,omitempty"`
	AllowedCharacters []string `json:"allowed_characters,omitempty"`
	Pattern           string   `json:"pattern,omitempty"`
	PatternMessage    string   `json:"pattern_message,omitempty"`
	Minimum           *float64 `json:"minimum,omitempty"`
	Maximum           *float64 `json:"maximum,omitempty"`
}

func NewFillFormData(form FillFormDataForm) FillFormData {
	return FillFormData{
		TypedData: TypedData{Type: DataTypeFillFormData},
		Form:      form,
	}
}

var _ authflow.Data = FillFormData{}

func (FillFormData) Data() {}

// NewFillFormDataForm builds the localized form for user. user carries the current values.
func NewFillFormDataForm(ctx context.Context, deps *authflow.Dependencies, form *config.AuthenticationFlowForm, user *model.User) (FillFormDataForm, error) {
	out := FillFormDataForm{
		Name:   form.Name,
		Fields: []FillFormDataField{},
	}

	for _, f := range form.Fields {
		pointer := formFieldKey(f)
		custom := findCustomAttribute(deps.Config, pointer)

		typ := fillFormFieldType(custom, pointer)
		label, err := formFieldLabel(ctx, deps, form.Name, pointer, custom)
		if err != nil {
			return FillFormDataForm{}, err
		}
		if custom != nil && custom.Type == config.CustomAttributeTypeBoolean {
			// The label may contain links, which are sanitized.
			label = string(htmlutil.SanitizeLinksHTML(label))
		}

		field := FillFormDataField{
			Key:      pointer,
			Type:     typ,
			Label:    label,
			Required: formFieldRequired(f),
			Value:    formFieldCurrentValue(user, pointer, custom),
		}
		if pointer == "/gender" {
			field.Suggestions = []string{"male", "female"}
		}

		field.Options, err = formFieldOptions(ctx, deps, pointer, custom)
		if err != nil {
			return FillFormDataForm{}, err
		}

		field.Constraints, err = formFieldConstraints(ctx, deps, pointer, custom)
		if err != nil {
			return FillFormDataForm{}, err
		}

		out.Fields = append(out.Fields, field)
	}

	return out, nil
}

// fillFormFieldType is the type table in the profile filling spec.
// custom is nil for a standard attribute.
func fillFormFieldType(custom *config.CustomAttributesAttributeConfig, pointer string) FillFormFieldType {
	if custom != nil {
		switch custom.Type {
		case config.CustomAttributeTypeString:
			return FillFormFieldTypeString
		case config.CustomAttributeTypeInteger:
			return FillFormFieldTypeInteger
		case config.CustomAttributeTypeNumber:
			return FillFormFieldTypeNumber
		case config.CustomAttributeTypeBoolean:
			return FillFormFieldTypeBoolean
		case config.CustomAttributeTypeEnum:
			return FillFormFieldTypeEnum
		case config.CustomAttributeTypePhoneNumber:
			return FillFormFieldTypePhoneNumber
		case config.CustomAttributeTypeEmail:
			return FillFormFieldTypeEmail
		case config.CustomAttributeTypeURL:
			return FillFormFieldTypeURL
		case config.CustomAttributeTypeCountryCode:
			return FillFormFieldTypeCountryCode
		}
		panic(fmt.Errorf("unknown custom attribute type: %v", custom.Type))
	}

	switch pointer {
	case "/name", "/given_name", "/family_name", "/middle_name", "/nickname", "/gender":
		return FillFormFieldTypeString
	case "/birthdate":
		return FillFormFieldTypeDate
	case "/zoneinfo":
		return FillFormFieldTypeTimezone
	case "/locale":
		return FillFormFieldTypeEnum
	case "/profile", "/picture", "/website":
		return FillFormFieldTypeURL
	case "/address":
		return FillFormFieldTypeAddress
	}
	// Config validation has ensured the attribute is fillable.
	panic(fmt.Errorf("unexpected form field: %v", pointer))
}

// formFieldLabel prefers the form's own label, so one attribute can be labelled differently per form.
func formFieldLabel(ctx context.Context, deps *authflow.Dependencies, formName string, pointer string, custom *config.CustomAttributesAttributeConfig) (string, error) {
	prefix := "standard-attribute-label-"
	if custom != nil {
		prefix = "custom-attribute-label-"
	}
	keys := []string{prefix + pointer}
	if formName != "" {
		keys = append([]string{"v2.page.fill-form." + formName + ".field-label-" + pointer}, keys...)
	}
	for _, key := range keys {
		ok, err := deps.Translations.HasKey(ctx, key)
		if err != nil {
			return "", err
		}
		if ok {
			return deps.Translations.RenderText(ctx, key, nil)
		}
	}
	return labelutil.Label(firstPointerToken(pointer)), nil
}

func firstPointerToken(pointer string) string {
	ptr, err := jsonpointer.Parse(pointer)
	if err != nil || len(ptr) == 0 {
		return pointer
	}
	return ptr[0]
}

func formFieldCurrentValue(user *model.User, pointer string, custom *config.CustomAttributesAttributeConfig) any {
	ptr, err := jsonpointer.Parse(pointer)
	if err != nil {
		return nil
	}
	attrs := user.StandardAttributes
	if custom != nil {
		attrs = user.CustomAttributes
	}
	v, err := ptr.Traverse(attrs)
	if err != nil {
		return nil
	}
	return v
}

func formFieldOptions(ctx context.Context, deps *authflow.Dependencies, pointer string, custom *config.CustomAttributesAttributeConfig) ([]FillFormDataFieldOption, error) {
	var options []FillFormDataFieldOption

	if custom != nil && custom.Type == config.CustomAttributeTypeEnum {
		for _, value := range custom.Enum {
			label, err := translatedOrDefault(ctx, deps, "custom-attribute-enum-label-"+pointer+"-"+value, labelutil.Label(value))
			if err != nil {
				return nil, err
			}
			options = append(options, FillFormDataFieldOption{Value: value, Label: label})
		}
	}

	if custom == nil && pointer == "/locale" {
		for _, tag := range deps.Config.Localization.SupportedLanguages {
			label, err := translatedOrDefault(ctx, deps, "language-"+tag, tag)
			if err != nil {
				return nil, err
			}
			options = append(options, FillFormDataFieldOption{Value: tag, Label: label})
		}
	}

	return options, nil
}

func translatedOrDefault(ctx context.Context, deps *authflow.Dependencies, key string, defaultText string) (string, error) {
	ok, err := deps.Translations.HasKey(ctx, key)
	if err != nil {
		return "", err
	}
	if !ok {
		return defaultText, nil
	}
	return deps.Translations.RenderText(ctx, key, nil)
}

func formFieldConstraints(ctx context.Context, deps *authflow.Dependencies, pointer string, custom *config.CustomAttributesAttributeConfig) (*FillFormDataFieldConstraints, error) {
	var stringConstraints *config.StringConstraints
	patternMessageKey := "standard-attribute-pattern-error-" + pointer
	if custom != nil {
		stringConstraints = custom.StringConstraints()
		patternMessageKey = "custom-attribute-pattern-error-" + pointer
	} else {
		stringConstraints = deps.Config.UserProfile.StandardAttributes.GetStringConstraints(pointer)
	}

	out := &FillFormDataFieldConstraints{}
	if stringConstraints != nil {
		out.MinLength = stringConstraints.MinLength
		out.MaxLength = stringConstraints.MaxLength
		for _, c := range stringConstraints.AllowedCharacters {
			out.AllowedCharacters = append(out.AllowedCharacters, string(c))
		}
		out.Pattern = stringConstraints.Pattern
		if stringConstraints.Pattern != "" {
			message, err := translatedOrDefault(ctx, deps, patternMessageKey, "")
			if err != nil {
				return nil, err
			}
			out.PatternMessage = message
		}
	}
	if custom != nil {
		out.Minimum = custom.Minimum
		out.Maximum = custom.Maximum
	}

	if out.isEmpty() {
		return nil, nil
	}
	return out, nil
}

func (c *FillFormDataFieldConstraints) isEmpty() bool {
	return c.MinLength == nil &&
		c.MaxLength == nil &&
		len(c.AllowedCharacters) == 0 &&
		c.Pattern == "" &&
		c.PatternMessage == "" &&
		c.Minimum == nil &&
		c.Maximum == nil
}
