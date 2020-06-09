package user

import (
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
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

type RawCommands struct {
	Store                         store
	Time                          time.Provider
	URLPrefix                     urlprefix.Provider
	TaskQueue                     async.Queue
	UserVerificationConfiguration *config.UserVerificationConfiguration
	WelcomeMessageProvider        WelcomeMessageProvider
}

func (c *RawCommands) Create(userID string, metadata map[string]interface{}) (*User, error) {
	now := c.Time.NowUTC()
	user := &User{
		ID:          userID,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: nil,
		Metadata:    metadata,
	}

	err := c.Store.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (c *RawCommands) AfterCreate(userModel *model.User, identities []*identity.Info) error {
	err := c.WelcomeMessageProvider.SendToIdentityInfos(identities)
	if err != nil {
		return err
	}

	if c.UserVerificationConfiguration.AutoSendOnSignup {
		c.enqueueSendVerificationCodeTasks(*userModel, identities)
	}

	return nil
}

func (c *RawCommands) enqueueSendVerificationCodeTasks(user model.User, identities []*identity.Info) {
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

func (c *RawCommands) UpdateMetadata(user *model.User, metadata map[string]interface{}) error {
	now := c.Time.NowUTC()
	if err := c.Store.UpdateMetadata(user.ID, metadata, now); err != nil {
		return err
	}

	user.Metadata = metadata
	return nil
}

func (c *RawCommands) UpdateLoginTime(user *model.User, lastLoginAt gotime.Time) error {
	if err := c.Store.UpdateLoginTime(user.ID, lastLoginAt); err != nil {
		return err
	}

	user.LastLoginAt = &lastLoginAt
	return nil
}
