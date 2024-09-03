package viewmodels

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func GetErrorJSON(apiError *apierrors.APIError) (errJSON map[string]interface{}) {
	errJSON = make(map[string]interface{})

	errJSONPtrs := resolveErrJSONPtrs(apiError)

	constructPtrJSON := func(ptr *ErrJSONPtr, eJSONError map[string]interface{}) map[string]interface{} {
		if ptr == nil {
			return nil
		}
		return map[string]interface{}{
			"error":  eJSONError,
			"hasMsg": ptr.HasMessage,
		}
	}
	eJSONError := parseAPIError(apiError)
	for _, ptr := range errJSONPtrs {
		errJSON[ptr.JSONPtr] = constructPtrJSON(ptr, eJSONError)
	}

	return
}

func parseAPIError(apiError *apierrors.APIError) (eJSONError map[string]interface{}) {
	b, err := json.Marshal(struct {
		Error *apierrors.APIError `json:"error"`
	}{apiError})
	if err != nil {
		panic(err)
	}

	var eJSON map[string]map[string]interface{}
	err = json.Unmarshal(b, &eJSON)
	if err != nil {
		panic(err)
	}
	eJSONError = eJSON["error"]
	return
}

type ErrJSONPtr struct {
	JSONPtr    string // json pointer to error
	HasMessage bool   // true if error should show message
}

// Resolve errJSONPtr from err
// it returns a slice because one error can display in multiple fields
func resolveErrJSONPtrs(apiError *apierrors.APIError) (errJSONPtrs []*ErrJSONPtr) {
	errJSONPtrs = []*ErrJSONPtr{}
	if apiError == nil {
		return
	}

	return
}
