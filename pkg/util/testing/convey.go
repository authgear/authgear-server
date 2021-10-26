package testing

import (
	"fmt"
	"reflect"
	"sort"

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

// ShouldEqualStringSliceWithoutOrder compares two string slice
// by considering them as string set
func ShouldEqualStringSliceWithoutOrder(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return "ShouldEqualStringSliceWithoutOrder receives only expected argument"
	}

	l, ok := actual.([]string)
	if !ok {
		return fmt.Sprintf("%v is not []string", actual)
	}

	r, ok := expected[0].([]string)
	if !ok {
		return fmt.Sprintf("%v is not []string", expected[0])
	}

	errMessage := func() string {
		return fmt.Sprintf(`Expected: '%v' Actual: '%v'`, l, r)
	}

	if len(l) != len(r) {
		return errMessage()
	}

	ll := make([]string, len(l))
	copy(ll, l)
	rr := make([]string, len(r))
	copy(rr, r)

	lll := sort.StringSlice(ll)
	lll.Sort()
	rrr := sort.StringSlice(rr)
	rrr.Sort()

	if !reflect.DeepEqual(lll, rrr) {
		return errMessage()
	}

	return ""
}
