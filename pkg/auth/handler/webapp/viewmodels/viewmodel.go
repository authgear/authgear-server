package viewmodels

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/authgear/authgear-server/pkg/lib/api/apierrors"
)

func sliceContains(slice []interface{}, value interface{}) bool {
	for _, v := range slice {
		if reflect.DeepEqual(v, value) {
			return true
		}
	}
	return false
}

func asAPIError(anyError interface{}) *apierrors.APIError {
	if err, ok := anyError.(error); ok {
		return apierrors.AsAPIError(err)
	}
	return nil
}

// Embed embeds the given struct s into data.
func Embed(data map[string]interface{}, s interface{}) {
	v := reflect.ValueOf(s)
	typ := v.Type()
	if typ.Kind() != reflect.Struct {
		panic(fmt.Errorf("webapp: expected struct but was %T", s))
	}
	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		structField := typ.Field(i)
		data[structField.Name] = v.Field(i).Interface()
	}
}

func EmbedForm(data map[string]interface{}, form url.Values) {
	for name := range form {
		data[name] = form.Get(name)
	}
}
