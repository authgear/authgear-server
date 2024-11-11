package customattrs

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
	"github.com/iawaknahc/jsonschema/pkg/jsonschema"

	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/customattrs"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/jsonpointerutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type UserQueries interface {
	GetRaw(ctx context.Context, userID string) (*user.User, error)
}

type UserStore interface {
	UpdateCustomAttributes(ctx context.Context, userID string, storageForm map[string]interface{}) error
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
	rootBuilder := validation.SchemaBuilder{}.
		Type(validation.TypeObject)

	properties := rootBuilder.Properties()

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

			var builder validation.SchemaBuilder
			builder, err = customAttr.ToSchemaBuilder()
			if err != nil {
				return
			}

			properties.Property(head, builder)
		}
	}

	schemaBytes, err := json.MarshalIndent(rootBuilder, "", "  ")
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

func (s *ServiceNoEvent) UpdateAllCustomAttributes(ctx context.Context, role accesscontrol.Role, userID string, reprForm map[string]interface{}) error {
	pointers := s.allPointers()
	return s.updateCustomAttributes(ctx, role, userID, pointers, reprForm)
}

func (s *ServiceNoEvent) UpdateCustomAttributesWithList(ctx context.Context, role accesscontrol.Role, userID string, l attrs.List) error {
	var pointers []string
	reprForm := make(map[string]interface{})

	for _, attr := range l {
		for _, c := range s.Config.CustomAttributes.Attributes {
			if attr.Pointer == c.Pointer {
				ptr, err := jsonpointer.Parse(c.Pointer)
				if err != nil {
					return err
				}

				pointers = append(pointers, c.Pointer)

				// nil means deletion
				if attr.Value == nil {
					continue
				}

				err = jsonpointerutil.AssignToJSONObject(ptr, reprForm, attr.Value)
				if err != nil {
					return err
				}
			}
		}
	}

	return s.updateCustomAttributes(ctx, role, userID, pointers, reprForm)
}

func (s *ServiceNoEvent) UpdateCustomAttributesWithForm(ctx context.Context, role accesscontrol.Role, userID string, form map[string]string) error {
	var pointers []string
	reprForm := make(map[string]interface{})

	for ptrStr, strRepr := range form {
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

	return s.updateCustomAttributes(ctx, role, userID, pointers, reprForm)
}

func (s *ServiceNoEvent) updateCustomAttributes(ctx context.Context, role accesscontrol.Role, userID string, pointers []string, reprForm map[string]interface{}) error {
	incoming := customattrs.T(reprForm)

	err := s.validate(pointers, incoming)
	if err != nil {
		return err
	}

	user, err := s.UserQueries.GetRaw(ctx, userID)
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

	err = s.UserStore.UpdateCustomAttributes(ctx, userID, storageForm)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceNoEvent) ReadCustomAttributesInStorageForm(
	ctx context.Context,
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

// Batch ReadCustomAttributesInStorageForm
func (s *ServiceNoEvent) ReadCustomAttributesInStorageFormForUsers(
	ctx context.Context,
	role accesscontrol.Role,
	userIDs []string,
	storageForms []map[string]interface{},
) (map[string]map[string]interface{}, error) {
	if len(userIDs) != len(storageForms) {
		panic("customattrs: expeceted same length of arguments")
	}

	customAttrsByUserID := map[string]map[string]interface{}{}

	for idx, userID := range userIDs {
		storageForm := storageForms[idx]
		c, err := s.ReadCustomAttributesInStorageForm(ctx, role, userID, storageForm)
		if err != nil {
			return nil, err
		}
		customAttrsByUserID[userID] = c
	}

	return customAttrsByUserID, nil
}
