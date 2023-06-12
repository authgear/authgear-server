package template

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/Masterminds/sprig"

	"github.com/authgear/authgear-server/pkg/util/messageformat"
)

func MakeTemplateFuncMap() map[string]interface{} {
	var templateFuncMap = sprig.HermeticHtmlFuncMap()
	templateFuncMap[messageformat.TemplateRuntimeFuncName] = messageformat.TemplateRuntimeFunc
	templateFuncMap["rfc3339"] = RFC3339
	templateFuncMap["ensureTime"] = EnsureTime
	templateFuncMap["isNil"] = IsNil
	templateFuncMap["showAttributeValue"] = ShowAttributeValue
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

var DefaultFuncMap = MakeTemplateFuncMap()
