package viewmodels

import (
	"net/url"
	"reflect"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/util/template"
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

func Embed(data map[string]interface{}, s interface{}) {
	template.Embed(data, s)
}

func EmbedForm(data map[string]interface{}, form url.Values) {
	for name := range form {
		data[name] = form.Get(name)
	}
}
