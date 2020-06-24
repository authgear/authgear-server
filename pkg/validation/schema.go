package validation

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/iawaknahc/jsonschema/pkg/jsonschema"
	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"
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

func (s *SimpleSchema) RegisterFormat(format string, checker jsonschemaformat.FormatChecker) {
	s.col.FormatChecker[format] = checker
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

func (s *MultipartSchema) DumpSchemaString(pretty bool) (schemaString string, err error) {
	schema := map[string]interface{}{
		"$defs": s.parts,
		"$ref":  jsonpointer.T([]string{"$defs", s.mainPartID}),
	}

	var schemaJSON []byte
	if pretty {
		schemaJSON, err = json.MarshalIndent(schema, "", "  ")
	} else {
		schemaJSON, err = json.Marshal(schema)
	}
	if err != nil {
		return
	}

	schemaString = string(schemaJSON)
	return
}

func (s *MultipartSchema) Instantiate() *MultipartSchema {
	if _, ok := s.parts[s.mainPartID]; !ok {
		panic(fmt.Sprintf("validaiton: main part '%s' is not added", s.mainPartID))
	}

	schemaString, err := s.DumpSchemaString(false)
	if err != nil {
		panic("validation: invalid JSON schema: " + err.Error())
	}

	// Do not forget the parts so that we can dump the schema later.
	// s.parts = nil
	s.col = jsonschema.NewCollection()
	s.col.AddSchema(strings.NewReader(schemaString), "")
	return s
}

func (s *MultipartSchema) RegisterFormat(format string, checker jsonschemaformat.FormatChecker) {
	if s.col == nil {
		panic("validation: JSON schema is not instantiated")
	}
	s.col.FormatChecker[format] = checker
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
		return &AggregatedError{Message: "invalid value", Errors: errs}
	}
	return nil
}
