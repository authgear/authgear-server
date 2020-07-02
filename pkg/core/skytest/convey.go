// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package skytest

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	coreReflect "github.com/authgear/authgear-server/pkg/core/reflect"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
	"github.com/smartystreets/goconvey/convey"
)

// ShouldEqualAPIError asserts the equality of skyerr.APIError
func ShouldEqualAPIError(actual interface{}, expected ...interface{}) string {
	if len(expected) < 1 || len(expected) > 2 {
		return fmt.Sprintf("ShouldEqualSkyError receives 1 to 2 arguments")
	}

	apiErr, ok := actual.(skyerr.APIError)
	if !ok {
		if err, ok := actual.(error); ok {
			apiErr = *skyerr.AsAPIError(err)
		} else {
			return fmt.Sprintf("%v is not convertible to skyerr.APIError", actual)
		}
	}

	// expected[0] is kind
	// expected[1] is info

	kind, ok := expected[0].(skyerr.Kind)
	if !ok {
		return fmt.Sprintf("%v is not skyerr.Kind", expected[0])
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
		Kind skyerr.Kind
		Info map[string]interface{}
	}

	return convey.ShouldResemble(APIError{apiErr.Kind, apiErr.Info}, APIError{kind, info})
}

// ShouldEqualStringSliceWithoutOrder compares two string slice
// by considering them as string set
func ShouldEqualStringSliceWithoutOrder(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("ShouldEqualStringSliceWithoutOrder receives only expected argument")
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

// ShouldEqualJSON asserts equality of two JSON bytes or strings by
// their key / value, regardless of the actual position of those
// key-value pairs
func ShouldEqualJSON(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("ShouldEqualJSON receives only one expected argument")
	}

	actualBytes, err := interfaceToByteSlice(actual)
	if err != nil {
		return fmt.Sprintf("%[1]v is %[1]T, not []byte or string", actual)
	}

	expectedBytes, err := interfaceToByteSlice(expected[0])
	if err != nil {
		return fmt.Sprintf("%[1]v is %[1]T, not []byte or string", expected[0])
	}

	var actualJSON, expectedJSON interface{}

	if err := json.Unmarshal(actualBytes, &actualJSON); err != nil {
		return fmt.Sprintf("invalid JSON of L.H.S.: %v; actual = \n%v", err, actual)
	}

	if err := json.Unmarshal(expectedBytes, &expectedJSON); err != nil {
		return fmt.Sprintf("invalid JSON of R.H.S.: %v; expected = \n%v", err, expected[0])
	}

	if !reflect.DeepEqual(actualJSON, expectedJSON) {
		return fmt.Sprintf(`Expected: '%s'
Actual:   '%s'`, prettyPrintJSONMap(expectedJSON), prettyPrintJSONMap(actualJSON))
	}

	return ""
}

func ShouldNonRecursiveDataDeepEqual(actual interface{}, expected ...interface{}) string {
	if len(expected) != 1 {
		return fmt.Sprintf("ShouldNonRecursiveDataDeepEqual expects only 1 argument")
	}
	if !coreReflect.NonRecursiveDataDeepEqual(actual, expected[0]) {
		return fmt.Sprintf(`Expected:
%+v

Actual:
%+v`, expected[0], actual)
	}

	return ""
}

func interfaceToByteSlice(i interface{}) ([]byte, error) {
	if b, ok := i.([]byte); ok {
		return b, nil
	}
	if s, ok := i.(string); ok {
		return []byte(s), nil
	}

	return nil, errors.New("cannot convert")
}

func prettyPrintJSONMap(i interface{}) []byte {
	b, _ := json.Marshal(&i)
	return b
}
