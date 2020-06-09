package user

import (
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	task "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type WelcomeMessageProvider interface {
	SendToIdentityInfos(infos []*identity.Info) error
}

type Commands struct {
	Store                         store
	Time                          time.Provider
	Hooks                         hook.Provider
	URLPrefix                     urlprefix.Provider
	TaskQueue                     async.Queue
	UserVerificationConfiguration *config.UserVerificationConfiguration
	WelcomeMessageProvider        WelcomeMessageProvider
}

func (c *Commands) Create(userID string, metadata map[string]interface{}, identities []*identity.Info) error {
	now := c.Time.NowUTC()
	user := &User{
		ID:          userID,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: nil,
	}

	err := c.Store.Create(user)
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

	err = c.WelcomeMessageProvider.SendToIdentityInfos(identities)
	if err != nil {
		return err
	}

	if c.UserVerificationConfiguration.AutoSendOnSignup {
		c.enqueueSendVerificationCodeTasks(*userModel, identities)
	}

	return nil
}

func (c *Commands) enqueueSendVerificationCodeTasks(user model.User, identities []*identity.Info) {
	for _, i := range identities {
		if i.Type != authn.IdentityTypeLoginID {
			continue
		}
		loginIDKey := i.Claims[identity.IdentityClaimLoginIDKey].(string)
		loginID := i.Claims[identity.IdentityClaimLoginIDValue].(string)

		for _, keyConfig := range c.UserVerificationConfiguration.LoginIDKeys {
			if keyConfig.Key == loginIDKey {
				c.TaskQueue.Enqueue(async.TaskSpec{
					Name: task.VerifyCodeSendTaskName,
					Param: task.VerifyCodeSendTaskParam{
						URLPrefix: c.URLPrefix.Value(),
						LoginID:   loginID,
						UserID:    user.ID,
					},
				})
			}
		}
	}
}

func (c *Commands) UpdateMetadata(user *User, metadata map[string]interface{}) error {
	user.UpdatedAt = c.Time.NowUTC()
	user.Metadata = metadata
	return c.Store.UpdateMetadata(user)
}

func (c *Commands) UpdateLoginTime(user *User, lastLoginAt gotime.Time) error {
	user.LastLoginAt = &lastLoginAt
	return c.Store.UpdateLoginTime(user)
}
