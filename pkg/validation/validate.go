package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/iawaknahc/jsonschema/pkg/jsonschema"
)

type SchemaValidator struct {
	Schema    *jsonschema.Collection
	Reference string
}

func (v *SchemaValidator) Validate(r io.Reader) error {
	return v.ValidateWithMessage(r, defaultErrorMessage)
}

func (v *SchemaValidator) ValidateWithMessage(r io.Reader, msg string) error {
	node, err := v.Schema.Apply(v.Reference, r)
	if err != nil {
		return fmt.Errorf("invalid JSON value: %w", err)
	}

	var errors []Error
	var traverseNode func(n *jsonschema.Node) bool
	traverseNode = func(n *jsonschema.Node) bool {
		if n.Valid {
			return true
		}

		hasInvalidChild := false
		for _, child := range n.Children {
			c := child
			if !traverseNode(&c) {
				hasInvalidChild = true
			}
		}

		if !hasInvalidChild {
			info, err := toJSONObject(n.Info)
			if err != nil {
				panic(fmt.Sprintf("validation: failed to marshal error info at %s: %s", n.KeywordLocation, err.Error()))
			}

			if len(info) == 0 && n.Keyword == "format" {
				if err, ok := n.Info.(error); ok {
					info = map[string]interface{}{"error": err.Error()}
				}
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

	if len(errors) != 0 {
		sort.Slice(errors, func(i, j int) bool {
			return errors[i].Location < errors[j].Location
		})
		return &AggregatedError{Message: msg, Errors: errors}
	}
	return nil
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
