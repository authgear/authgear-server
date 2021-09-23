package user

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type Commands struct {
	*RawCommands
	Events       EventService
	Verification VerificationService
}

func (c *Commands) AfterCreate(
	user *User,
	identities []*identity.Info,
	authenticators []*authenticator.Info,
	isAdminAPI bool,
	webhookState string,
) error {
	isVerified, err := c.Verification.IsUserVerified(identities)
	if err != nil {
		return err
	}

	userModel := newUserModel(user, identities, authenticators, isVerified)
	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}

	events := []event.Payload{
		&blocking.UserPreCreateBlockingEventPayload{
			User:       *userModel,
			Identities: identityModels,
			AdminAPI:   isAdminAPI,
			OAuthState: webhookState,
		},
		&nonblocking.UserCreatedEventPayload{
			User:       *userModel,
			Identities: identityModels,
			AdminAPI:   isAdminAPI,
		},
	}

	for _, e := range events {
		if err := c.Events.DispatchEvent(e); err != nil {
			return err
		}
	}

	err = c.RawCommands.AfterCreate(userModel, identities)
	if err != nil {
		return err
	}

	return nil
}
