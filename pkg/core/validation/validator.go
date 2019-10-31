package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

// Validator is a collection of schemas.
type Validator struct {
	RootSchemaID        string
	definitions         map[string]interface{}
	schemaLoader        *gojsonschema.SchemaLoader
	compiledSchemaCache sync.Map
}

func NewValidator(rootSchemaID string) *Validator {
	return &Validator{
		RootSchemaID:        rootSchemaID,
		definitions:         map[string]interface{}{},
		schemaLoader:        gojsonschema.NewSchemaLoader(),
		compiledSchemaCache: sync.Map{},
	}
}

type subschema struct {
	id     string
	key    string
	schema map[string]interface{}
}

func (v *Validator) AddSchemaFragments(schemaStrings ...string) error {
	for _, schemaString := range schemaStrings {
		schemas, err := v.parseSchemaFragment(schemaString)
		if err != nil {
			return err
		}

		for _, subschema := range schemas {
			if _, ok := v.definitions[subschema.key]; ok {
				return fmt.Errorf("duplicate definitions key: %s", subschema.key)
			}
			v.definitions[subschema.key] = subschema.schema
		}
	}

	b, err := json.Marshal(map[string]interface{}{
		"$id":         v.RootSchemaID,
		"definitions": v.definitions,
	})
	if err != nil {
		return err
	}

	err = v.schemaLoader.AddSchemas(gojsonschema.NewBytesLoader(b))
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator) ValidateGoValue(schemaID string, value interface{}) error {
	loader := gojsonschema.NewGoLoader(value)
	return v.validateWithLoader(schemaID, loader)
}

func (v *Validator) ParseReader(schemaID string, r io.Reader, value interface{}) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	loader := gojsonschema.NewBytesLoader(b)
	err = v.validateWithLoader(schemaID, loader)
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewReader(b)).Decode(value)
	if err != nil {
		return err
	}
	return nil
}

func (v *Validator) validateWithLoader(schemaID string, loader gojsonschema.JSONLoader) error {
	schema, err := v.getSchema(schemaID)
	if err != nil {
		return err
	}
	result, err := schema.Validate(loader)
	if err != nil {
		return err
	}
	if !result.Valid() {
		return ConvertErrors(result.Errors())
	}
	return nil
}

func (v *Validator) getSchema(schemaID string) (schema *gojsonschema.Schema, err error) {
	absoluteID := v.RootSchemaID + schemaID
	cached, ok := v.compiledSchemaCache.Load(absoluteID)
	if !ok {
		schema, err = v.schemaLoader.Compile(gojsonschema.NewReferenceLoader(absoluteID))
		if err != nil {
			return nil, err
		}
		cached, _ = v.compiledSchemaCache.LoadOrStore(absoluteID, schema)
	}
	schema = cached.(*gojsonschema.Schema)
	return schema, nil
}

func (v *Validator) parseSchemaFragment(subschemaString string) ([]subschema, error) {
	var subschemas []subschema
	var schemaMap map[string]interface{}
	err := json.Unmarshal([]byte(subschemaString), &schemaMap)
	if err != nil {
		return nil, err
	}

	if id, ok := schemaMap["$id"].(string); ok {
		return []subschema{subschema{
			key:    strings.TrimPrefix(id, "#"),
			id:     id,
			schema: schemaMap,
		}}, nil
	}

	for key, subschemaVal := range schemaMap {
		subschemaMap, ok := subschemaVal.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("schema must be an object")
		}

		id, ok := subschemaMap["$id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid id of schema: %s", key)
		}

		subschemas = append(subschemas, subschema{
			key:    key,
			id:     id,
			schema: subschemaMap,
		})
	}

	return subschemas, nil
}
