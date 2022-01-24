package template

import (
	"time"

	"github.com/Masterminds/sprig"
	messageformat "github.com/iawaknahc/gomessageformat"
)

func MakeTemplateFuncMap() map[string]interface{} {
	var templateFuncMap = sprig.HermeticHtmlFuncMap()
	templateFuncMap[messageformat.TemplateRuntimeFuncName] = messageformat.TemplateRuntimeFunc
	templateFuncMap["rfc3339"] = RFC3339
	templateFuncMap["ensureTime"] = EnsureTime
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

var DefaultFuncMap = MakeTemplateFuncMap()
