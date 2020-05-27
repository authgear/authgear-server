package user

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	task "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type WelcomeMessageProvider interface {
	SendToIdentityInfos(infos []*identity.Info) error
}

type Commands struct {
	AuthInfos                     authinfo.Store
	UserProfiles                  userprofile.Store
	Time                          time.Provider
	Hooks                         hook.Provider
	URLPrefix                     urlprefix.Provider
	TaskQueue                     async.Queue
	UserVerificationConfiguration *config.UserVerificationConfiguration
	WelcomeMessageProvider        WelcomeMessageProvider
}

func (c *Commands) Create(userID string, metadata map[string]interface{}, identities []*identity.Info) error {
	now := c.Time.NowUTC()
	authInfo := &authinfo.AuthInfo{
		ID:          userID,
		VerifyInfo:  map[string]bool{},
		LastLoginAt: &now,
	}

	err := c.AuthInfos.CreateAuth(authInfo)
	if err != nil {
		return err
	}

	userProfile, err := c.UserProfiles.CreateUserProfile(authInfo.ID, metadata)
	if err != nil {
		return err
	}

	user := newUser(now, authInfo, &userProfile, identities)
	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, i.ToModel())
	}
	err = c.Hooks.DispatchEvent(
		event.UserCreateEvent{
			User:       *user,
			Identities: identityModels,
		},
		user,
	)
	if err != nil {
		return err
	}

	err = c.WelcomeMessageProvider.SendToIdentityInfos(identities)
	if err != nil {
		return err
	}

	if c.UserVerificationConfiguration.AutoSendOnSignup {
		c.enqueueSendVerificationCodeTasks(*user, identities)
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
