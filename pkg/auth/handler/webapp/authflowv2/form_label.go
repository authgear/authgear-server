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
