package customattrs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/iawaknahc/jsonschema/pkg/jsonschema"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/jsonpointerutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type Service struct {
	Config *config.CustomAttributesConfig
}

func (s *Service) FromStorageForm(storageForm map[string]interface{}) (T, error) {
	out := make(T)
	for _, c := range s.Config.Attributes {
		ptr, err := jsonpointer.Parse(c.Pointer)
		if err != nil {
			return nil, err
		}

		if val, ok := storageForm[c.ID]; ok {
			err = jsonpointerutil.AssignToJSONObject(ptr, out, val)
			if err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}

func (s *Service) GenerateSchemaString(pointers []string) (schemaStr string, err error) {
	properties := make(map[string]interface{})

	for _, ptrStr := range pointers {
		for _, customAttr := range s.Config.Attributes {
			if ptrStr != customAttr.Pointer {
				continue
			}

			var ptr jsonpointer.T
			ptr, err = jsonpointer.Parse(ptrStr)
			if err != nil {
				return
			}
			head := ptr[0]

			var schema map[string]interface{}
			schema, err = CustomAttributeConfigToSchema(customAttr)
			if err != nil {
				return
			}

			properties[head] = schema
		}
	}

	schemaObj := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	schemaBytes, err := json.MarshalIndent(schemaObj, "", "  ")
	if err != nil {
		return
	}

	schemaStr = string(schemaBytes)
	return
}

func (s *Service) Validate(pointers []string, input T) error {
	schemaStr, err := s.GenerateSchemaString(pointers)
	if err != nil {
		return err
	}

	col := jsonschema.NewCollection()
	err = col.AddSchema(strings.NewReader(schemaStr), "")
	if err != nil {
		return err
	}

	validator := &validation.SchemaValidator{
		Schema: col,
	}

	err = validator.ValidateValue(map[string]interface{}(input))
	if err != nil {
		return err
	}

	return nil
}

func CustomAttributeConfigToSchema(customAttr *config.CustomAttributesAttributeConfig) (schema map[string]interface{}, err error) {
	schema = make(map[string]interface{})

	switch customAttr.Type {
	case config.CustomAttributeTypeString:
		schema["type"] = "string"
	case config.CustomAttributeTypeNumber:
		schema["type"] = "number"
		if customAttr.Minimum != nil {
			schema["minimum"] = *customAttr.Minimum
		}
		if customAttr.Maximum != nil {
			schema["maximum"] = *customAttr.Maximum
		}
	case config.CustomAttributeTypeInteger:
		schema["type"] = "integer"
		if customAttr.Minimum != nil {
			schema["minimum"] = int64(*customAttr.Minimum)
		}
		if customAttr.Maximum != nil {
			schema["maximum"] = int64(*customAttr.Maximum)
		}
	case config.CustomAttributeTypeEnum:
		schema["type"] = "string"
		schema["enum"] = customAttr.Enum
	case config.CustomAttributeTypePhoneNumber:
		schema["type"] = "string"
		schema["format"] = "phone"
	case config.CustomAttributeTypeEmail:
		schema["type"] = "string"
		schema["format"] = "email"
	case config.CustomAttributeTypeURL:
		schema["type"] = "string"
		schema["format"] = "uri"
	case config.CustomAttributeTypeAlpha2:
		schema["type"] = "string"
		schema["format"] = "iso3166-1-alpha-2"
	default:
		err = fmt.Errorf("unknown type: %v", customAttr.Type)
	}

	return
}
