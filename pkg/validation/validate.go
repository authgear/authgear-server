package validation

import (
	"encoding/json"
	"fmt"
	"github.com/iawaknahc/jsonschema/pkg/jsonschema"
	"io"
	"sort"
)

func validateSchema(col *jsonschema.Collection, r io.Reader) ([]Error, error) {
	node, err := col.Apply("", r)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON value: %w", err)
	}

	var errors []Error
	var traverseNode func(n *jsonschema.Node) bool
	traverseNode = func(n *jsonschema.Node) bool {
		if n.Valid {
			return true
		}

		hasInvalidChild := false
		for _, child := range n.Children {
			if !traverseNode(&child) {
				hasInvalidChild = true
			}
		}

		if !hasInvalidChild {
			info, err := toJSONObject(n.Info)
			if err != nil {
				panic(fmt.Sprintf("validation: failed to marshal error info at %s: %s", n.KeywordLocation, err.Error()))
			}
			errors = append(errors, Error{
				Location: n.InstanceLocation.String(),
				Keyword:  n.Keyword,
				Info:     info,
			})
		}

		return false
	}
	traverseNode(node)

	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Location < errors[j].Location
	})
	return errors, nil
}

func toJSONObject(data interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var obj map[string]interface{}
	err = json.Unmarshal(jsonData, &obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}
