package template

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/sprig"

	"github.com/authgear/authgear-server/pkg/util/messageformat"
)

const (
	templateTranslationMessageTemplateName = "__translation_message.html"
)

type tpl interface {
	ExecuteTemplate(wr io.Writer, name string, data any) error
}

func MakeTemplateFuncMap(t tpl) map[string]interface{} {
	templateFuncMap := makeTemplateFuncMap()
	templateFuncMap["include"] = makeInclude(t)
	templateFuncMap["translate"] = makeTranslate(t)
	templateFuncMap["trimHTML"] = trimHTML
	return templateFuncMap
}

func makeTemplateFuncMap() map[string]interface{} {
	var templateFuncMap = sprig.HermeticHtmlFuncMap()
	templateFuncMap[messageformat.TemplateRuntimeFuncName] = messageformat.TemplateRuntimeFunc
	templateFuncMap["rfc3339"] = RFC3339
	templateFuncMap["ensureTime"] = EnsureTime
	templateFuncMap["isNil"] = IsNil
	templateFuncMap["showAttributeValue"] = ShowAttributeValue
	templateFuncMap["htmlattr"] = HTMLAttr
	return templateFuncMap
}

func RFC3339(date interface{}) interface{} {
	switch date := date.(type) {
	case *time.Time:
		return date.UTC().Format(time.RFC3339)
	case time.Time:
		return date.UTC().Format(time.RFC3339)
	default:
		return "INVALID_DATE"
	}
}

func EnsureTime(anyValue interface{}) interface{} {
	switch anyValue := anyValue.(type) {
	case *time.Time:
		return anyValue
	case time.Time:
		return anyValue
	case string:
		t, err := time.Parse(time.RFC3339, anyValue)
		if err != nil {
			panic(err)
		}
		return t
	case *string:
		t, err := time.Parse(time.RFC3339, *anyValue)
		if err != nil {
			panic(err)
		}
		return t
	default:
		return anyValue
	}
}

func IsNil(v interface{}) bool {
	return v == nil ||
		(reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func ShowAttributeValue(v interface{}) string {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		if !value.IsNil() {
			return ShowAttributeValue(reflect.ValueOf(v).Elem().Interface())
		}
		return ""
	}

	switch v := v.(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)

	}
}

func HTMLAttr(v string) template.HTMLAttr {
	// Ignore gosec error because the app developer can actually write any template
	// But we should be careful that do not pass any user input to this function
	return template.HTMLAttr(v) // nolint:gosec
}

func makeInclude(t tpl) func(tplName string, data any) (template.HTML, error) {
	return func(
		tplName string,
		data any,
	) (template.HTML, error) {
		buf := &bytes.Buffer{}
		err := t.ExecuteTemplate(buf, tplName, data)
		// Ignore gosec error because the app developer can actually write any template
		// But we should be careful that do not pass any user input to this function
		html := template.HTML(buf.String()) // nolint:gosec
		return html, err
	}
}

// `translate` is intended for `include` a translation message but wrapped it
// in a span and set its translation key with data attribute
// In theory it can be used with resources other than translation, but take your
// own risks
func makeTranslate(t tpl) func(tranlsationKey string, data any) (template.HTML, error) {
	include := makeInclude(t)
	return func(
		tranlsationKey string,
		data any,
	) (template.HTML, error) {
		included, err := include(tranlsationKey, data)
		if err != nil {
			return template.HTML(""), err
		}
		buf := &bytes.Buffer{}
		d := make(map[string]interface{})
		d["Key"] = tranlsationKey
		d["Value"] = included
		err = t.ExecuteTemplate(buf, templateTranslationMessageTemplateName, d)
		// Ignore gosec error because the app developer can actually write any template
		// But we should be careful that do not pass any user input to this function
		html := template.HTML(buf.String()) // nolint:gosec
		return html, err
	}
}

func trimHTML(input interface{}) (interface{}) {
	switch input := input.(type) {
	case string:
		return strings.TrimSpace(input)
	case template.HTML:
		// `Masterminds/sprig`'s `trimAll` cannot handle html type, so we need to convert it to string first
		return template.HTML(strings.TrimSpace(string(input)))
	default:
		return ""
	}
}
