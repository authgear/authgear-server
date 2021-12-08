package customattrs

import (
	"encoding/json"
	"strings"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/iawaknahc/jsonschema/pkg/jsonschema"

	"github.com/authgear/authgear-server/pkg/lib/authn/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/jsonpointerutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type UserStore interface {
	UpdateCustomAttributes(userID string, storageForm map[string]interface{}) error
}

type Service struct {
	Config    *config.UserProfileConfig
	UserStore UserStore
}

func (s *Service) FromStorageForm(storageForm map[string]interface{}) (customattrs.T, error) {
	out := make(customattrs.T)
	for _, c := range s.Config.CustomAttributes.Attributes {
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

func (s *Service) ToStorageForm(t customattrs.T) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	for _, c := range s.Config.CustomAttributes.Attributes {
		ptr, err := jsonpointer.Parse(c.Pointer)
		if err != nil {
			return nil, err
		}

		if val, err := ptr.Traverse(t); err == nil {
			out[c.ID] = val
		}
	}
	return out, nil
}

func (s *Service) GenerateSchemaString(pointers []string) (schemaStr string, err error) {
	properties := make(map[string]interface{})

	for _, ptrStr := range pointers {
		for _, customAttr := range s.Config.CustomAttributes.Attributes {
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
			schema, err = customAttr.ToJSONSchema()
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

func (s *Service) Validate(pointers []string, input customattrs.T) error {
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

func (s *Service) allPointers() (out []string) {
	for _, c := range s.Config.CustomAttributes.Attributes {
		out = append(out, c.Pointer)
	}
	return
}

func (s *Service) UpdateCustomAttributes(userID string, reprForm map[string]interface{}) error {
	customAttrs := customattrs.T(reprForm)
	pointers := s.allPointers()

	err := s.Validate(pointers, customAttrs)
	if err != nil {
		return err
	}

	storageForm, err := s.ToStorageForm(customAttrs)
	if err != nil {
		return err
	}

	err = s.UserStore.UpdateCustomAttributes(userID, storageForm)
	if err != nil {
		return err
	}

	return nil
}
