package customattrs

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/attrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type Service struct {
	Config         *config.UserProfileConfig
	ServiceNoEvent *ServiceNoEvent
	Events         EventService
}

func (s *Service) UpdateAllCustomAttributes(ctx context.Context, role accesscontrol.Role, userID string, reprForm map[string]interface{}) error {
	err := s.ServiceNoEvent.UpdateAllCustomAttributes(ctx, role, userID, reprForm)
	if err != nil {
		return err
	}

	err = s.dispatchEvents(ctx, role, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateCustomAttributesWithList(ctx context.Context, role accesscontrol.Role, userID string, attrs attrs.List) error {
	err := s.ServiceNoEvent.UpdateCustomAttributesWithList(ctx, role, userID, attrs)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateCustomAttributesWithForm(ctx context.Context, role accesscontrol.Role, userID string, jsonPointerMap map[string]string) error {
	err := s.ServiceNoEvent.UpdateCustomAttributesWithForm(ctx, role, userID, jsonPointerMap)
	if err != nil {
		return err
	}

	err = s.dispatchEvents(ctx, role, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) dispatchEvents(ctx context.Context, role accesscontrol.Role, userID string) (err error) {
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
		err = s.Events.DispatchEventOnCommit(ctx, eventPayload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) ReadCustomAttributesInStorageForm(
	ctx context.Context,
	role accesscontrol.Role,
	userID string,
	storageForm map[string]interface{},
) (map[string]interface{}, error) {
	return s.ServiceNoEvent.ReadCustomAttributesInStorageForm(ctx, role, userID, storageForm)
}
