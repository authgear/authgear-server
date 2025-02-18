package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/iawaknahc/jsonschema/pkg/jsonschema"
)

var errJSONSyntaxErrorOfEmptyInput error

func init() {
	var unimportant interface{}
	// errJSONSyntaxErrorOfEmptyInput is a *json.SyntaxError with msg set.
	// We have to do this because *json.SyntaxError.msg is private,
	// we cannot initialize msg.
	errJSONSyntaxErrorOfEmptyInput = json.Unmarshal(nil, &unimportant)
}

type SchemaValidator struct {
	Schema    *jsonschema.Collection
	Reference string
}

func (v *SchemaValidator) Parse(ctx context.Context, r io.Reader, value interface{}) error {
	return v.ParseWithMessage(ctx, r, defaultErrorMessage, value)
}

func (v *SchemaValidator) ParseWithMessage(ctx context.Context, r io.Reader, msg string, value interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	err = v.ValidateWithMessage(ctx, bytes.NewReader(data), msg)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, value)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func (v *SchemaValidator) ParseJSONRawMessage(ctx context.Context, msg json.RawMessage, value interface{}) error {
	return v.ParseWithMessage(ctx, bytes.NewReader(msg), defaultErrorMessage, value)
}

func (v *SchemaValidator) ValidateValue(ctx context.Context, value interface{}) error {
	return v.ValidateValueWithMessage(ctx, value, defaultErrorMessage)
}

func (v *SchemaValidator) ValidateValueWithMessage(ctx context.Context, value interface{}, msg string) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return v.ValidateWithMessage(ctx, bytes.NewReader(data), msg)
}

func (v *SchemaValidator) Validate(ctx context.Context, r io.Reader) error {
	return v.ValidateWithMessage(ctx, r, defaultErrorMessage)
}

func (v *SchemaValidator) ValidateWithMessage(ctx context.Context, r io.Reader, msg string) error {
	node, err := v.Schema.Apply(ctx, v.Reference, r)
	if err != nil {
		// It is observed that json.NewDecoder.Decode and json.Unmarshal
		// returns different error with the input is empty.
		// json.Unmarshal returns *json.SyntaxError while
		// json.NewDecoder.Decode returns io.EOF
		// https://go.dev/play/p/WHEtDYzJKTo
		// So we convert the io.EOF here.
		if errors.Is(err, io.EOF) {
			return errJSONSyntaxErrorOfEmptyInput
		}

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
				panic(fmt.Sprintf("validation: failed to marshal error info at %s: %s", n.Verbose().KeywordLocation, err.Error()))
			}

			keyword := n.Keyword
			if len(info) == 0 && keyword == "format" {
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
				Keyword:  keyword,
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
