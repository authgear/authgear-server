package authflowv2

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow/declarative"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/labelutil"
)

type FormFieldOptionViewModel struct {
	Value string
	Label string
}

// formFieldLabel resolves the label of the field with key in the form named formName.
func formFieldLabel(t viewmodels.TranslationsCompat, formName string, key string) (string, error) {
	if formName != "" {
		k := "v2.page.fill-form." + formName + ".field-label-" + key
		ok, err := t.HasKey(k)
		if err != nil {
			return "", err
		}
		if ok {
			return t.RenderText(k, nil)
		}
	}

	// The only field kind is user_profile, whose key is its pointer.
	prefix := "custom-attribute-label-"
	if config.IsStandardAttributePointer(key) {
		prefix = "standard-attribute-label-"
	}
	return translatedOrDefault(t, prefix+key, labelutil.Label(firstPointerToken(key)))
}

// formFieldOptions resolves the label of each option of the field with key.
func formFieldOptions(t viewmodels.TranslationsCompat, key string, options []declarative.FillFormDataFieldOption) ([]FormFieldOptionViewModel, error) {
	if options == nil {
		return nil, nil
	}
	out := make([]FormFieldOptionViewModel, 0, len(options))
	for _, o := range options {
		var label string
		var err error
		if key == "/locale" {
			label, err = translatedOrDefault(t, "language-"+o.Value, o.Value)
		} else {
			label, err = translatedOrDefault(t, "custom-attribute-enum-label-"+key+"-"+o.Value, labelutil.Label(o.Value))
		}
		if err != nil {
			return nil, err
		}
		out = append(out, FormFieldOptionViewModel{Value: o.Value, Label: label})
	}
	return out, nil
}

// formFieldPlaceholder resolves the placeholder of the field with key in the form named formName.
// Only a named form has placeholders, and only on fields that are typed into.
func formFieldPlaceholder(t viewmodels.TranslationsCompat, formName string, key string, fieldType declarative.FillFormFieldType) (string, error) {
	if formName == "" || !hasPlaceholder(fieldType) {
		return "", nil
	}
	return translatedOrDefault(t, "v2.page.fill-form."+formName+".field-placeholder-"+key, "")
}

func hasPlaceholder(fieldType declarative.FillFormFieldType) bool {
	switch fieldType {
	case declarative.FillFormFieldTypeString,
		declarative.FillFormFieldTypeEmail,
		declarative.FillFormFieldTypeURL,
		declarative.FillFormFieldTypeInteger,
		declarative.FillFormFieldTypeNumber,
		declarative.FillFormFieldTypePhoneNumber:
		return true
	default:
		return false
	}
}

// formPageTextKeys resolves the translation keys of the title and description of a fill_form page.
// An empty description of a named form is kept, so it hides the description.
func formPageTextKeys(t viewmodels.TranslationsCompat, formName string) (titleKey string, descriptionKey string, err error) {
	titleKey = "v2.page.fill-form.default.title"
	descriptionKey = "v2.page.fill-form.default.description"
	if formName == "" {
		return
	}
	named := "v2.page.fill-form." + formName + ".title"
	ok, err := t.HasKey(named)
	if err != nil {
		return "", "", err
	}
	if ok {
		titleKey = named
	}
	named = "v2.page.fill-form." + formName + ".description"
	ok, err = t.DefinesKey(named)
	if err != nil {
		return "", "", err
	}
	if ok {
		descriptionKey = named
	}
	return
}

func translatedOrDefault(t viewmodels.TranslationsCompat, key string, defaultText string) (string, error) {
	ok, err := t.HasKey(key)
	if err != nil {
		return "", err
	}
	if !ok {
		return defaultText, nil
	}
	return t.RenderText(key, nil)
}

func firstPointerToken(pointer string) string {
	ptr, err := jsonpointer.Parse(pointer)
	if err != nil || len(ptr) == 0 {
		return pointer
	}
	return ptr[0]
}
