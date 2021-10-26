package testing

import (
	"fmt"

	"github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

// ShouldEqualAPIError asserts the equality of apierrors.APIError
func ShouldEqualAPIError(actual interface{}, expected ...interface{}) string {
	if len(expected) < 1 || len(expected) > 2 {
		return "ShouldEqualSkyError receives 1 to 2 arguments"
	}

	apiErr, ok := actual.(apierrors.APIError)
	if !ok {
		if err, ok := actual.(error); ok {
			apiErr = *apierrors.AsAPIError(err)
		} else {
			return fmt.Sprintf("%v is not convertible to apierrors.APIError", actual)
		}
	}

	// expected[0] is kind
	// expected[1] is info

	kind, ok := expected[0].(apierrors.Kind)
	if !ok {
		return fmt.Sprintf("%v is not apierrors.Kind", expected[0])
	}

	var info map[string]interface{}
	if len(expected) == 2 {
		info, ok = expected[1].(map[string]interface{})
		if !ok {
			return fmt.Sprintf("%v is not map[string]interface{}", expected[1])
		}
	} else {
		info = apiErr.Info
	}

	type APIError struct {
		Kind apierrors.Kind
		Info map[string]interface{}
	}

	return convey.ShouldResemble(APIError{apiErr.Kind, apiErr.Info}, APIError{kind, info})
}
