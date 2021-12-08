package facade

import (
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type CustomAttributesService interface {
	ReadCustomAttributesInStorageForm(role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
	UpdateAllCustomAttributes(role accesscontrol.Role, userID string, customAttrs map[string]interface{}) error
}

type CustomAttributesFacade struct {
	CustomAttributes CustomAttributesService
}

func (f *CustomAttributesFacade) ReadCustomAttributesInStorageForm(
	role accesscontrol.Role,
	userID string,
	storageForm map[string]interface{},
) (map[string]interface{}, error) {
	return f.CustomAttributes.ReadCustomAttributesInStorageForm(role, userID, storageForm)
}

func (f *CustomAttributesFacade) UpdateAllCustomAttributes(role accesscontrol.Role, userID string, customAttrs map[string]interface{}) error {
	return f.CustomAttributes.UpdateAllCustomAttributes(role, userID, customAttrs)
}
