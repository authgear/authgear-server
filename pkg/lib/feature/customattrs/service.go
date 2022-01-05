package customattrs

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type Service struct {
	Config         *config.UserProfileConfig
	ServiceNoEvent *ServiceNoEvent
	Events         EventService
}

func (s *Service) UpdateAllCustomAttributes(role accesscontrol.Role, userID string, reprForm map[string]interface{}) error {
	err := s.ServiceNoEvent.UpdateAllCustomAttributes(role, userID, reprForm)
	if err != nil {
		return err
	}

	err = s.dispatchEvents(role, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateCustomAttributesWithJSONPointerMap(role accesscontrol.Role, userID string, jsonPointerMap map[string]string) error {
	err := s.ServiceNoEvent.UpdateCustomAttributesWithJSONPointerMap(role, userID, jsonPointerMap)
	if err != nil {
		return err
	}

	err = s.dispatchEvents(role, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) dispatchEvents(role accesscontrol.Role, userID string) (err error) {
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
		err = s.Events.DispatchEvent(eventPayload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ReadCustomAttributesInStorageForm(
	role accesscontrol.Role,
	userID string,
	storageForm map[string]interface{},
) (map[string]interface{}, error) {
	return s.ServiceNoEvent.ReadCustomAttributesInStorageForm(role, userID, storageForm)
}
