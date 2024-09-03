package viewmodels

import (
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func GetErrorJSON(apiError *apierrors.APIError) (errJSON map[string]map[string]interface{}) {
	errJSON = make(map[string]map[string]interface{})

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
	ptrs := resolvePasswordInput(apiError)
	errJSONPtrs = append(errJSONPtrs, ptrs...)
	return
}

// TODO: Or should pass this resolver per-page?
// ref pkg/auth/handler/webapp/authflowv2/viewmodels/password_input_error.go
// nolint: gocognit
func resolvePasswordInput(apiError *apierrors.APIError) (errJSONPtrs []*ErrJSONPtr) {
	errJSONPtrs = []*ErrJSONPtr{}
	addPtr := func(ptr *ErrJSONPtr) {
		errJSONPtrs = append(errJSONPtrs, ptr)
	}
	addPasswordInputPtr := func(hasMsg bool) {
		addPtr(&ErrJSONPtr{
			JSONPtr:    "password-input",
			HasMessage: hasMsg,
		})
	}
	addConfirmPasswordInputPtr := func(hasMsg bool) {
		addPtr(&ErrJSONPtr{
			JSONPtr:    "confirm-password-input",
			HasMessage: hasMsg,
		})
	}
	if apiError == nil {
		return
	}

	switch apiError.Reason {
	case "InvalidCredentials":
		addPasswordInputPtr(true)
	case "PasswordPolicyViolated":
		addPasswordInputPtr(true)
		addConfirmPasswordInputPtr(false)
	case "NewPasswordTypo":
		addConfirmPasswordInputPtr(true)
	case "ValidationFailed":
		for _, causes := range apiError.Info["causes"].([]interface{}) {
			if cause, ok := causes.(map[string]interface{}); ok {
				if kind, ok := cause["kind"].(string); ok {
					if kind == "required" {
						if details, ok := cause["details"].(map[string]interface{}); ok {
							if missing, ok := details["missing"].([]interface{}); ok {
								if SliceContains(missing, "x_password") {
									addPasswordInputPtr(true)
								} else if SliceContains(missing, "x_new_password") {
									addPasswordInputPtr(true)
								} else if SliceContains(missing, "x_confirm_password") {
									addConfirmPasswordInputPtr(true)
								}
							}
						}
					}
				}
			}
		}

	}

	return
}
