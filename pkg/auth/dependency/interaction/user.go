package interaction

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

type userProvider struct {
	AuthInfos                     authinfo.Store
	UserProfiles                  userprofile.Store
	Time                          time.Provider
	Hooks                         hook.Provider
	URLPrefix                     urlprefix.Provider
	TaskQueue                     async.Queue
	UserVerificationConfiguration *config.UserVerificationConfiguration
	WelcomeMessageProvider        WelcomeMessageProvider
}

func (p *userProvider) Create(userID string, metadata map[string]interface{}, identities []*identity.Info) error {
	now := p.Time.NowUTC()
	authInfo := &authinfo.AuthInfo{
		ID:          userID,
		VerifyInfo:  map[string]bool{},
		LastLoginAt: &now,
	}

	err := p.AuthInfos.CreateAuth(authInfo)
	if err != nil {
		return err
	}

	userProfile, err := p.UserProfiles.CreateUserProfile(authInfo.ID, metadata)
	if err != nil {
		return err
	}

	user := model.NewUser(*authInfo, userProfile)
	var identityModels []model.Identity
	for _, i := range identities {
		identityModels = append(identityModels, model.Identity{
			Type:   string(i.Type),
			Claims: i.Claims,
		})
	}
	err = p.Hooks.DispatchEvent(
		event.UserCreateEvent{
			User:       user,
			Identities: identityModels,
		},
		&user,
	)
	if err != nil {
		return err
	}

	err = p.WelcomeMessageProvider.SendToIdentityInfos(identities)
	if err != nil {
		return err
	}

	if p.UserVerificationConfiguration.AutoSendOnSignup {
		p.enqueueSendVerificationCodeTasks(user, identities)
	}

	return nil
}

func (p *userProvider) Get(userID string) (*model.User, error) {
	var authInfo authinfo.AuthInfo
	err := p.AuthInfos.GetAuth(userID, &authInfo)
	if err != nil {
		return nil, err
	}

	userProfile, err := p.UserProfiles.GetUserProfile(userID)
	if err != nil {
		return nil, err
	}

	u := model.NewUser(authInfo, userProfile)
	return &u, nil
}

func (p *userProvider) enqueueSendVerificationCodeTasks(user model.User, identities []*identity.Info) {
	for _, i := range identities {
		if i.Type != authn.IdentityTypeLoginID {
			continue
		}
		loginIDKey := i.Claims[identity.IdentityClaimLoginIDKey].(string)
		loginID := i.Claims[identity.IdentityClaimLoginIDValue].(string)

		for _, keyConfig := range p.UserVerificationConfiguration.LoginIDKeys {
			if keyConfig.Key == loginIDKey {
				p.TaskQueue.Enqueue(async.TaskSpec{
					Name: task.VerifyCodeSendTaskName,
					Param: task.VerifyCodeSendTaskParam{
						URLPrefix: p.URLPrefix.Value(),
						LoginID:   loginID,
						UserID:    user.ID,
					},
				})
			}
		}
	}
}
