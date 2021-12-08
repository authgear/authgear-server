package facade

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type StandardAttributesService interface {
	UpdateStandardAttributes(role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
	DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
}

type CustomAttributesService interface {
	ReadCustomAttributesInStorageForm(role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
	UpdateAllCustomAttributes(role accesscontrol.Role, userID string, customAttrs map[string]interface{}) error
}

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type UserProfileFacade struct {
	StandardAttributes StandardAttributesService
	CustomAttributes   CustomAttributesService
	Events             EventService
}

func (f *UserProfileFacade) DeriveStandardAttributes(role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error) {
	return f.StandardAttributes.DeriveStandardAttributes(role, userID, updatedAt, attrs)
}

func (f *UserProfileFacade) ReadCustomAttributesInStorageForm(
	role accesscontrol.Role,
	userID string,
	storageForm map[string]interface{},
) (map[string]interface{}, error) {
	return f.CustomAttributes.ReadCustomAttributesInStorageForm(role, userID, storageForm)
}

func (f *UserProfileFacade) UpdateUserProfile(
	role accesscontrol.Role,
	userID string,
	stdAttrs map[string]interface{},
	customAttrs map[string]interface{},
) error {
	err := f.StandardAttributes.UpdateStandardAttributes(role, userID, stdAttrs)
	if err != nil {
		return err
	}

	err = f.CustomAttributes.UpdateAllCustomAttributes(role, userID, customAttrs)
	if err != nil {
		return err
	}

	eventPayloads := []event.Payload{
		&blocking.UserProfilePreUpdateBlockingEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			},
			AdminAPI: role == accesscontrol.RoleGreatest,
		},
		&nonblocking.UserProfileUpdatedEventPayload{
			UserRef: model.UserRef{
				Meta: model.Meta{
					ID: userID,
				},
			},
			AdminAPI: role == accesscontrol.RoleGreatest,
		},
	}

	for _, eventPayload := range eventPayloads {
		err = f.Events.DispatchEvent(eventPayload)
		if err != nil {
			return err
		}
	}

	return nil
}
