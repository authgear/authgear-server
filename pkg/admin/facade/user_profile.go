package facade

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type StandardAttributesService interface {
	UpdateStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, stdAttrs map[string]interface{}) error
	DeriveStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error)
}

type CustomAttributesService interface {
	ReadCustomAttributesInStorageForm(ctx context.Context, role accesscontrol.Role, userID string, storageForm map[string]interface{}) (map[string]interface{}, error)
	UpdateAllCustomAttributes(ctx context.Context, role accesscontrol.Role, userID string, customAttrs map[string]interface{}) error
}

type EventService interface {
	DispatchEventOnCommit(ctx context.Context, payload event.Payload) error
}

type UserProfileFacade struct {
	User               UserService
	StandardAttributes StandardAttributesService
	CustomAttributes   CustomAttributesService
	Events             EventService
}

func (f *UserProfileFacade) DeriveStandardAttributes(ctx context.Context, role accesscontrol.Role, userID string, updatedAt time.Time, attrs map[string]interface{}) (map[string]interface{}, error) {
	return f.StandardAttributes.DeriveStandardAttributes(ctx, role, userID, updatedAt, attrs)
}

func (f *UserProfileFacade) ReadCustomAttributesInStorageForm(ctx context.Context,
	role accesscontrol.Role,
	userID string,
	storageForm map[string]interface{},
) (map[string]interface{}, error) {
	return f.CustomAttributes.ReadCustomAttributesInStorageForm(ctx, role, userID, storageForm)
}

func (f *UserProfileFacade) UpdateUserProfile(ctx context.Context,
	role accesscontrol.Role,
	userID string,
	stdAttrs map[string]interface{},
	customAttrs map[string]interface{},
) (err error) {
	updated := false
	err = f.User.CheckUserAnonymized(ctx, userID)
	if err != nil {
		return err
	}

	if stdAttrs != nil {
		updated = true
		err = f.StandardAttributes.UpdateStandardAttributes(ctx, role, userID, stdAttrs)
		if err != nil {
			return
		}
	}

	if customAttrs != nil {
		updated = true
		err = f.CustomAttributes.UpdateAllCustomAttributes(ctx, role, userID, customAttrs)
		if err != nil {
			return
		}
	}

	if updated {
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
			err = f.Events.DispatchEventOnCommit(ctx, eventPayload)
			if err != nil {
				return
			}
		}
	}

	return
}
