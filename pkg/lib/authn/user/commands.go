package user

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

type EventService interface {
	DispatchEventOnCommit(payload event.Payload) error
}

type Commands struct {
	*RawCommands
	RawQueries         *RawQueries
	Events             EventService
	Verification       VerificationService
	UserProfileConfig  *config.UserProfileConfig
	StandardAttributes StandardAttributesService
	CustomAttributes   CustomAttributesService
	Web3               Web3Service
	RolesAndGroups     RolesAndGroupsService
}

func (c *Commands) AfterCreate(
	user *User,
	identities []*identity.Info,
	authenticators []*authenticator.Info,
	isAdminAPI bool,
) error {
	isVerified, err := c.Verification.IsUserVerified(identities)
	if err != nil {
		return err
	}

	stdAttrs, err := c.StandardAttributes.DeriveStandardAttributes(accesscontrol.RoleGreatest, user.ID, user.UpdatedAt, user.StandardAttributes)
	if err != nil {
		return err
	}

	customAttrs, err := c.CustomAttributes.ReadCustomAttributesInStorageForm(accesscontrol.RoleGreatest, user.ID, user.CustomAttributes)
	if err != nil {
		return err
	}

	web3Info, err := c.Web3.GetWeb3Info(identities)
	if err != nil {
		return err
	}

	roles, err := c.RolesAndGroups.ListRolesByUserID(user.ID)
	if err != nil {
		return err
	}
	roleKeys := make([]string, len(roles))
	for i, v := range roles {
		roleKeys[i] = v.Key
	}

	groups, err := c.RolesAndGroups.ListGroupsByUserID(user.ID)
	if err != nil {
		return err
	}
	groupKeys := make([]string, len(groups))
	for i, v := range groups {
		groupKeys[i] = v.Key
	}

	userModel := newUserModel(user, identities, authenticators, isVerified, stdAttrs, customAttrs, web3Info, roleKeys, groupKeys)
	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}

	events := []event.Payload{
		&blocking.UserPreCreateBlockingEventPayload{
			UserRef:    *user.ToRef(),
			Identities: identityModels,
			AdminAPI:   isAdminAPI,
		},
		&nonblocking.UserCreatedEventPayload{
			UserRef:    *user.ToRef(),
			Identities: identityModels,
			AdminAPI:   isAdminAPI,
		},
	}

	for _, e := range events {
		if err := c.Events.DispatchEventOnCommit(e); err != nil {
			return err
		}
	}

	err = c.RawCommands.AfterCreate(userModel, identities)
	if err != nil {
		return err
	}

	return nil
}
