package skytest

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// ShouldEqualJSON asserts eqaulity of two JSON bytes or strings by
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
