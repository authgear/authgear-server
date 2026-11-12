package declarative

import (
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
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
	Required bool              `json:"required"`
	// Value is omitted only when nil. A false is kept.
	Value       any                           `json:"value,omitempty"`
	Options     []FillFormDataFieldOption     `json:"options,omitempty"`
	Suggestions []string                      `json:"suggestions,omitempty"`
	Constraints *FillFormDataFieldConstraints `json:"constraints,omitempty"`
}

type FillFormDataFieldOption struct {
	Value string `json:"value"`
}

type FillFormDataFieldConstraints struct {
	MinLength         *int     `json:"min_length,omitempty"`
	MaxLength         *int     `json:"max_length,omitempty"`
	AllowedCharacters []string `json:"allowed_characters,omitempty"`
	Pattern           string   `json:"pattern,omitempty"`
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

// NewFillFormDataForm builds the form for user. user carries the current values.
func NewFillFormDataForm(deps *authflow.Dependencies, form *config.AuthenticationFlowForm, user *model.User) FillFormDataForm {
	out := FillFormDataForm{
		Name:   form.Name,
		Fields: []FillFormDataField{},
	}

	for _, f := range form.Fields {
		pointer := formFieldKey(f)
		custom := findCustomAttribute(deps.Config, pointer)

		field := FillFormDataField{
			Key:      pointer,
			Type:     fillFormFieldType(custom, pointer),
			Required: formFieldRequired(f),
			Value:    formFieldCurrentValue(user, pointer, custom),
			Options:     formFieldOptions(deps, pointer, custom),
			Constraints: formFieldConstraints(deps, pointer, custom),
		}
		if pointer == "/gender" {
			field.Suggestions = []string{"male", "female"}
		}

		out.Fields = append(out.Fields, field)
	}

	return out
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

func formFieldOptions(deps *authflow.Dependencies, pointer string, custom *config.CustomAttributesAttributeConfig) []FillFormDataFieldOption {
	var options []FillFormDataFieldOption

	if custom != nil && custom.Type == config.CustomAttributeTypeEnum {
		for _, value := range custom.Enum {
			options = append(options, FillFormDataFieldOption{Value: value})
		}
	}

	if custom == nil && pointer == "/locale" {
		for _, tag := range deps.Config.Localization.SupportedLanguages {
			options = append(options, FillFormDataFieldOption{Value: tag})
		}
	}

	return options
}

func formFieldConstraints(deps *authflow.Dependencies, pointer string, custom *config.CustomAttributesAttributeConfig) *FillFormDataFieldConstraints {
	var stringConstraints *config.StringConstraints
	if custom != nil {
		stringConstraints = custom.StringConstraints()
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
	}
	if custom != nil {
		out.Minimum = custom.Minimum
		out.Maximum = custom.Maximum
	}

	if out.isEmpty() {
		return nil
	}
	return out
}

func (c *FillFormDataFieldConstraints) isEmpty() bool {
	return c.MinLength == nil &&
		c.MaxLength == nil &&
		len(c.AllowedCharacters) == 0 &&
		c.Pattern == "" &&
		c.Minimum == nil &&
		c.Maximum == nil
}
