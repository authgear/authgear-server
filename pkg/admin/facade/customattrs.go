package facade

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/customattrs"
)

type CustomAttributesService interface {
	FromStorageForm(storageForm map[string]interface{}) (customattrs.T, error)
	UpdateCustomAttributes(id string, customAttrs map[string]interface{}) error
}

type CustomAttributesFacade struct {
	CustomAttributes CustomAttributesService
}

func (f *CustomAttributesFacade) FromStorageForm(storageForm map[string]interface{}) (customattrs.T, error) {
	return f.CustomAttributes.FromStorageForm(storageForm)
}

func (f *CustomAttributesFacade) UpdateCustomAttributes(id string, customAttrs map[string]interface{}) error {
	return f.CustomAttributes.UpdateCustomAttributes(id, customAttrs)
}
