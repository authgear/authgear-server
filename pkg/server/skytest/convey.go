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

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// ShouldEqualSkyError asserts the equality of skyerr.Error
func ShouldEqualSkyError(actual interface{}, expected ...interface{}) string {
	if len(expected) < 1 || len(expected) > 3 {
		return fmt.Sprintf("ShouldEqualSkyError receives 1 to 3 arguments")
	}

	lhs, ok := actual.(skyerr.Error)
	if !ok {
		return fmt.Sprintf("%v is not skyerr.Error", actual)
	}

	// expected[0] is code
	// expected[1] is message
	// expected[2] is info

	code, ok := expected[0].(skyerr.ErrorCode)
	if !ok {
		return fmt.Sprintf("%v is not skyerr.ErrorCode", expected[0])
	}

	message := ""
	if len(expected) >= 2 {
		message, ok = expected[1].(string)
		if !ok {
			return fmt.Sprintf("%v is not message", expected[1])
		}
	}

	var info map[string]interface{}
	if len(expected) == 3 {
		info, ok = expected[2].(map[string]interface{})
		if !ok {
			return fmt.Sprintf("%v is not info", expected[2])
		}
	}

	rhs := skyerr.NewErrorWithInfo(code, message, info)
	if !reflect.DeepEqual(lhs, rhs) {
		return fmt.Sprintf(`Expected: '%v' Actual: '%v'`, lhs, rhs)
	}

	return ""
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
