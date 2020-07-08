package webapp

import (
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

func asAPIError(anyError interface{}) *skyerr.APIError {
	if apiError, ok := anyError.(*skyerr.APIError); ok {
		return apiError
	}
	if err, ok := anyError.(error); ok {
		return skyerr.AsAPIError(err)
	}
	return nil
}
