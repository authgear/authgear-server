package customattrs

import (
	"encoding/json"
	"strings"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/iawaknahc/jsonschema/pkg/jsonschema"

	"github.com/authgear/authgear-server/pkg/lib/authn/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/jsonpointerutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type UserQueries interface {
	GetRaw(userID string) (*user.User, error)
}

type UserStore interface {
	UpdateCustomAttributes(userID string, storageForm map[string]interface{}) error
}

type ServiceNoEvent struct {
	Config      *config.UserProfileConfig
	UserQueries UserQueries
	UserStore   UserStore
}

func (s *ServiceNoEvent) fromStorageForm(storageForm map[string]interface{}) (customattrs.T, error) {
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

func (s *ServiceNoEvent) toStorageForm(t customattrs.T) (map[string]interface{}, error) {
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

func (s *ServiceNoEvent) generateSchemaString(pointers []string) (schemaStr string, err error) {
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

func (s *ServiceNoEvent) validate(pointers []string, input customattrs.T) error {
	schemaStr, err := s.generateSchemaString(pointers)
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

func (s *ServiceNoEvent) allPointers() (out []string) {
	for _, c := range s.Config.CustomAttributes.Attributes {
		out = append(out, c.Pointer)
	}
	return
}

func (s *ServiceNoEvent) UpdateAllCustomAttributes(role accesscontrol.Role, userID string, reprForm map[string]interface{}) error {
	pointers := s.allPointers()
	return s.updateCustomAttributes(role, userID, pointers, reprForm)
}

func (s *ServiceNoEvent) UpdateCustomAttributesWithJSONPointerMap(role accesscontrol.Role, userID string, jsonPointerMap map[string]string) error {
	var pointers []string
	reprForm := make(map[string]interface{})

	for ptrStr, strRepr := range jsonPointerMap {
		for _, c := range s.Config.CustomAttributes.Attributes {
			if ptrStr == c.Pointer {
				ptr, err := jsonpointer.Parse(c.Pointer)
				if err != nil {
					return err
				}

				pointers = append(pointers, c.Pointer)

				// Empty string means deletion
				if strRepr == "" {
					continue
				}

				// In case of error, use the string representation as value.
				if val, err := c.ParseString(strRepr); err == nil {
					err = jsonpointerutil.AssignToJSONObject(ptr, reprForm, val)
					if err != nil {
						return err
					}
				} else {
					err = jsonpointerutil.AssignToJSONObject(ptr, reprForm, strRepr)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return s.updateCustomAttributes(role, userID, pointers, reprForm)
}

func (s *ServiceNoEvent) updateCustomAttributes(role accesscontrol.Role, userID string, pointers []string, reprForm map[string]interface{}) error {
	incoming := customattrs.T(reprForm)

	err := s.validate(pointers, incoming)
	if err != nil {
		return err
	}

	user, err := s.UserQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	original, err := s.fromStorageForm(user.CustomAttributes)
	if err != nil {
		return err
	}

	accessControl := s.Config.CustomAttributes.GetAccessControl()

	updated, err := original.Update(accessControl, role, pointers, incoming)
	if err != nil {
		return err
	}

	storageForm, err := s.toStorageForm(updated)
	if err != nil {
		return err
	}

	err = s.UserStore.UpdateCustomAttributes(userID, storageForm)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceNoEvent) ReadCustomAttributesInStorageForm(
	role accesscontrol.Role,
	userID string,
	storageForm map[string]interface{},
) (map[string]interface{}, error) {
	accessControl := s.Config.CustomAttributes.GetAccessControl()
	repr, err := s.fromStorageForm(storageForm)
	if err != nil {
		return nil, err
	}
	repr = repr.ReadWithAccessControl(accessControl, role)
	return repr.ToMap(), nil
}
