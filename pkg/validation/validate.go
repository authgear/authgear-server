package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/iawaknahc/jsonschema/pkg/jsonschema"
)

type SchemaValidator struct {
	Schema    *jsonschema.Collection
	Reference string
}

func (v *SchemaValidator) Parse(r io.Reader, value interface{}) error {
	return v.ParseWithMessage(r, defaultErrorMessage, value)
}

func (v *SchemaValidator) ParseWithMessage(r io.Reader, msg string, value interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	err = v.ValidateWithMessage(bytes.NewReader(data), msg)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func (v *SchemaValidator) ValidateValue(value interface{}) error {
	return v.ValidateValueWithMessage(value, defaultErrorMessage)
}

func (v *SchemaValidator) ValidateValueWithMessage(value interface{}, msg string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return v.ValidateWithMessage(bytes.NewReader(data), msg)
}

func (v *SchemaValidator) Validate(r io.Reader) error {
	return v.ValidateWithMessage(r, defaultErrorMessage)
}

func (v *SchemaValidator) ValidateWithMessage(r io.Reader, msg string) error {
	node, err := v.Schema.Apply(v.Reference, r)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
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
					info = map[string]interface{}{
						"error": err.Error(),
					}
				} else if info == nil {
					info = map[string]interface{}{}
				}
				info["format"] = n.Annotation.(string)
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
