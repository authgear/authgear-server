package template

import (
	"fmt"
)

// MakeMap creates a map with the given key value pairs.
func MakeMap(pairs ...interface{}) map[string]interface{} {
	length := len(pairs)
	if length%2 == 1 {
		panic(fmt.Errorf("template: the length of the argument must be even: %v", length))
	}

	out := make(map[string]interface{})
	for i := 0; i < length; i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			panic(fmt.Errorf("template: unexpected key type: %T %v", pairs[i], pairs[i]))
		}
		out[key] = pairs[i+1]
	}

	return out
}
