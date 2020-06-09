package user

import (
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type Commands struct {
	Raw   *RawCommands
	Time  time.Provider
	Hooks hook.Provider
}

func (c *Commands) Create(userID string, metadata map[string]interface{}, identities []*identity.Info) error {
	user, err := c.Raw.Create(userID, metadata)
	if err != nil {
		return err
	}

	userModel := newUserModel(user, identities)
	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}
	err = c.Hooks.DispatchEvent(
		event.UserCreateEvent{
			User:       *userModel,
			Identities: identityModels,
		},
		userModel,
	)
	if err != nil {
		return err
	}

	err = c.Raw.AfterCreate(userModel, identities)
	if err != nil {
		return err
	}

	return nil
}

func (c *Commands) UpdateMetadata(user *model.User, metadata map[string]interface{}) error {
	// TODO(webhook): dispatch update metadata event
	return c.Raw.UpdateMetadata(user, metadata)
}

func (c *Commands) UpdateLoginTime(user *model.User, lastLoginAt gotime.Time) error {
	return c.Raw.UpdateLoginTime(user, lastLoginAt)
}
