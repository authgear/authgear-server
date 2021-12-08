package facade

import (
	"time"

	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type StandardAttributesService interface {
	UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
	DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
}

type StandardAttributesFacade struct {
	StandardAttributes StandardAttributesService
}

func (f *StandardAttributesFacade) DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error) {
	return f.StandardAttributes.DeriveStandardAttributes(role, userID, updatedAt, attrs)
}

func (f *StandardAttributesFacade) UpdateStandardAttributes(role accesscontrol.Role, id string, stdAttrs map[string]interface{}) error {
	return f.StandardAttributes.UpdateStandardAttributes(role, id, stdAttrs)
}
