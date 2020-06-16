package validation

import (
	"encoding/json"
	"fmt"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/iawaknahc/jsonschema/pkg/jsonschema"
	"io"
	"strings"
)

type SimpleSchema struct {
	col *jsonschema.Collection
}

func NewSimpleSchema(schema string) *SimpleSchema {
	col := jsonschema.NewCollection()
	col.AddSchema(strings.NewReader(schema), "")
	return &SimpleSchema{
		col: col,
	}
}

func (s *SimpleSchema) ValidateReader(r io.Reader) error {
	return convertErrors(validateSchema(s.col, r, ""))
}

type MultipartSchema struct {
	mainPartID string
	parts      map[string]interface{}
	col        *jsonschema.Collection
}

func NewMultipartSchema(mainPartID string) *MultipartSchema {
	return &MultipartSchema{
		mainPartID: mainPartID,
		parts:      map[string]interface{}{},
		col:        nil,
	}
}

func (s *MultipartSchema) Add(partID string, schema string) *MultipartSchema {
	if s.col != nil {
		panic("validation: cannot add part when schema is already instantiated")
	}
	var schemaObj interface{}
	if err := json.Unmarshal([]byte(schema), &schemaObj); err != nil {
		panic(fmt.Sprintf("validation: invalid schema part '%s': %s", partID, err))
	}
	s.parts[partID] = schemaObj
	return s
}

func (s *MultipartSchema) Instantiate() *MultipartSchema {
	if _, ok := s.parts[s.mainPartID]; !ok {
		panic(fmt.Sprintf("validaiton: main part '%s' is not added", s.mainPartID))
	}
	schema := map[string]interface{}{
		"$defs": s.parts,
		"$ref":  jsonpointer.T([]string{"$defs", s.mainPartID}),
	}
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		panic("validation: invalid JSON schema: " + err.Error())
	}
	s.parts = nil

	s.col = jsonschema.NewCollection()
	s.col.AddSchema(strings.NewReader(string(schemaJSON)), "")

	return s
}

func (s *MultipartSchema) ValidateReader(r io.Reader) error {
	if s.col == nil {
		panic("validation: JSON schema is not instantiated")
	}
	return convertErrors(validateSchema(s.col, r, ""))
}

func (s *MultipartSchema) ValidateReaderByPart(r io.Reader, partID string) error {
	if s.col == nil {
		panic("validation: JSON schema is not instantiated")
	}
	return convertErrors(validateSchema(s.col, r, jsonpointer.T([]string{"$defs", partID}).Fragment()))
}

func convertErrors(errs []Error, err error) error {
	if err != nil {
		return fmt.Errorf("failed to validate JSON: %w", err)
	}
	if len(errs) != 0 {
		return &AggregatedError{Errors: errs}
	}
	return nil
}
