package user

import (
	gotime "time"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type HookProvider interface {
	DispatchEvent(payload event.Payload) error
}

type Commands struct {
	Raw          *RawCommands
	Hooks        HookProvider
	Verification VerificationService
}

func (c *Commands) Create(userID string) (*User, error) {
	return c.Raw.Create(userID)
}

func (c *Commands) AfterCreate(
	user *User,
	identities []*identity.Info,
) error {
	isVerified, err := c.Verification.IsUserVerified(identities)
	if err != nil {
		return err
	}

	userModel := newUserModel(user, identities, isVerified)
	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}
	err = c.Hooks.DispatchEvent(&event.UserCreateEvent{
		User:       *userModel,
		Identities: identityModels,
	})
	if err != nil {
		return err
	}

	err = c.Raw.AfterCreate(userModel, identities)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) UpdateLoginTime(userID string, loginAt gotime.Time) error {
	return c.Raw.UpdateLoginTime(userID, loginAt)
}

func (c *Commands) UpdateDisabledStatus(userID string, isDisabled bool, reason *string) error {
	return c.Raw.UpdateDisabledStatus(userID, isDisabled, reason)
}

func (c *Commands) Delete(userID string) error {
	return c.Raw.Delete(userID)
}
