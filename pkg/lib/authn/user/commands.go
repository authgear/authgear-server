package user

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type Commands struct {
	*RawCommands
	RawQueries        *RawQueries
	Events            EventService
	Verification      VerificationService
	UserProfileConfig *config.UserProfileConfig
}

func (c *Commands) PopulateStandardAttributes(userID string, iden *identity.Info) error {
	user, err := c.RawQueries.GetRaw(userID)
	if err != nil {
		return err
	}

	stdAttrsFromIden := stdattrs.T(iden.Claims).NonIdentityAware()
	originalStdAttrs := stdattrs.T(user.StandardAttributes)
	stdAttrs := originalStdAttrs.MergedWith(stdAttrsFromIden)

	err = c.RawCommands.UpdateStandardAttributes(userID, stdAttrs.ToClaims())
	if err != nil {
		return err
	}

	return nil
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

	stdAttrs, err := c.Verification.DeriveStandardAttributes(user.ID, user.StandardAttributes)
	if err != nil {
		return err
	}

	userModel := newUserModel(user, identities, authenticators, isVerified, stdAttrs)
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
