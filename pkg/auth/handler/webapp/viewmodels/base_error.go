package viewmodels

import "github.com/authgear/authgear-server/pkg/api/apierrors"


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
