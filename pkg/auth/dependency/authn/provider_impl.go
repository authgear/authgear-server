package authn

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/audit"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	task "github.com/skygeario/skygear-server/pkg/auth/task/spec"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/validation"
)

type ProviderImpl struct {
	PasswordChecker               *audit.PasswordChecker
	LoginIDChecker                loginid.LoginIDChecker
	IdentityProvider              principal.IdentityProvider
	TimeProvider                  coreTime.Provider
	AuthInfoStore                 authinfo.Store
	UserProfileStore              userprofile.Store
	PasswordProvider              password.Provider
	HookProvider                  hook.Provider
	WelcomeEmailConfiguration     *config.WelcomeEmailConfiguration
	UserVerificationConfiguration *config.UserVerificationConfiguration
	AuthConfiguration             *config.AuthConfiguration
	URLPrefixProvider             urlprefix.Provider
}

func (p *ProviderImpl) CreateUserWithLoginIDs(
	loginIDs []loginid.LoginID,
	plainPassword string,
	metadata map[string]interface{},
	onUserDuplicate model.OnUserDuplicate,
) (authInfo *authinfo.AuthInfo, userProfile *userprofile.UserProfile, firstPrincipal principal.Principal, tasks []async.TaskSpec, err error) {
	err = p.validateCreateUserWithLoginIDs(
		loginIDs,
		onUserDuplicate,
	)
	if err != nil {
		return
	}

	err = p.PasswordChecker.ValidatePassword(audit.ValidatePasswordPayload{
		PlainPassword: plainPassword,
	})
	if err != nil {
		return
	}

	existingPrincipals, err := p.findExistingPrincipalsWithProposedLoginIDs(loginIDs)
	if err != nil {
		return
	}

	if len(existingPrincipals) > 0 && onUserDuplicate == model.OnUserDuplicateAbort {
		err = password.ErrLoginIDAlreadyUsed
		return
	}

	now := p.TimeProvider.NowUTC()
	authInfoS := authinfo.NewAuthInfo()
	authInfo = &authInfoS
	authInfo.LastLoginAt = &now

	err = p.AuthInfoStore.CreateAuth(authInfo)
	if err != nil {
		return
	}

	userProfileS, err := p.UserProfileStore.CreateUserProfile(authInfo.ID, metadata)
	if err != nil {
		return
	}
	userProfile = &userProfileS

	principals, err := p.createPrincipalsWithProposedLoginIDs(authInfo.ID, plainPassword, loginIDs)
	if err != nil {
		return
	}
	firstPrincipal = principals[0]

	user := model.NewUser(*authInfo, *userProfile)
	identities := []model.Identity{}
	for _, principal := range principals {
		identity := model.NewIdentity(p.IdentityProvider, principal)
		identities = append(identities, identity)
	}

	err = p.HookProvider.DispatchEvent(
		event.UserCreateEvent{
			User:       user,
			Identities: identities,
		},
		&user,
	)
	if err != nil {
		return
	}

	if p.WelcomeEmailConfiguration.Enabled {
		tasks = append(tasks, p.generateSendWelcomeEmailTasks(user, loginIDs)...)
	}

	if p.UserVerificationConfiguration.AutoSendOnSignup {
		tasks = append(tasks, p.generateSendVerificationCodeTasks(user, loginIDs)...)
	}

	return
}

func (p *ProviderImpl) validateCreateUserWithLoginIDs(loginIDs []loginid.LoginID, onUserDuplicate model.OnUserDuplicate) (err error) {
	var causes []validation.ErrorCause

	if !model.IsAllowedOnUserDuplicate(
		false,
		p.AuthConfiguration.OnUserDuplicateAllowCreate,
		onUserDuplicate,
	) {
		causes = append(causes, validation.ErrorCause{
			Kind:    validation.ErrorGeneral,
			Pointer: "/on_user_duplicate",
			Message: "on_user_duplicate is not allowed",
		})
	}

	seen := map[string]struct{}{}

	for i, loginID := range loginIDs {
		if _, found := seen[loginID.Value]; found {
			causes = append(causes, validation.ErrorCause{
				Kind:    validation.ErrorGeneral,
				Pointer: fmt.Sprintf("/login_ids/%d/value", i),
				Message: "duplicated login ID",
			})
		}
		seen[loginID.Value] = struct{}{}
	}

	if err := p.LoginIDChecker.Validate(loginIDs); err != nil {
		if cs := validation.ErrorCauses(err); len(cs) > 0 {
			for i := range cs {
				cs[i].Pointer = fmt.Sprintf("/login_ids%s", cs[i].Pointer)
			}
			causes = append(causes, cs...)
		}
	}

	if len(causes) > 0 {
		err = validation.NewValidationFailed("invalid request body", causes)
		return
	}

	return nil
}

func (p *ProviderImpl) findExistingPrincipalsWithProposedLoginIDs(loginIDs []loginid.LoginID) (principals []principal.Principal, err error) {
	// Find out all login IDs that are of type email.
	var emails []string
	for _, loginID := range loginIDs {
		if p.LoginIDChecker.CheckType(loginID.Key, metadata.Email) {
			emails = append(emails, loginID.Value)
		}
	}

	// For each email, find out all principals.
	var ps []principal.Principal
	for _, email := range emails {
		ps, err = p.IdentityProvider.ListPrincipalsByClaim("email", email)
		if err != nil {
			return
		}
		principals = append(principals, ps...)
	}

	realm := password.DefaultRealm
	var filteredPrincipals []principal.Principal
	for _, p := range principals {
		if passwordPrincipal, ok := p.(*password.Principal); ok && passwordPrincipal.Realm != realm {
			continue
		}
		filteredPrincipals = append(filteredPrincipals, p)
	}

	principals = filteredPrincipals
	return
}

func (p *ProviderImpl) createPrincipalsWithProposedLoginIDs(userID string, plainPassword string, loginIDs []loginid.LoginID) (principals []principal.Principal, err error) {
	realm := password.DefaultRealm
	passwordPrincipals, err := p.PasswordProvider.CreatePrincipalsByLoginID(
		userID,
		plainPassword,
		loginIDs,
		realm,
	)
	if err != nil {
		return
	}

	for _, principal := range passwordPrincipals {
		principals = append(principals, principal)
	}
	return
}

func (p *ProviderImpl) generateSendWelcomeEmailTasks(user model.User, loginIDs []loginid.LoginID) (tasks []async.TaskSpec) {
	supportedLoginIDs := []loginid.LoginID{}
	for _, loginID := range loginIDs {
		if p.LoginIDChecker.CheckType(loginID.Key, metadata.Email) {
			supportedLoginIDs = append(supportedLoginIDs, loginID)
		}
	}

	var destinationLoginIDs []loginid.LoginID
	if p.WelcomeEmailConfiguration.Destination == config.WelcomeEmailDestinationAll {
		destinationLoginIDs = supportedLoginIDs
	} else if p.WelcomeEmailConfiguration.Destination == config.WelcomeEmailDestinationFirst {
		if len(supportedLoginIDs) > 0 {
			destinationLoginIDs = supportedLoginIDs[:1]
		}
	}

	for _, loginID := range destinationLoginIDs {
		email := loginID.Value
		tasks = append(tasks, async.TaskSpec{
			Name: task.WelcomeEmailSendTaskName,
			Param: task.WelcomeEmailSendTaskParam{
				URLPrefix: p.URLPrefixProvider.Value(),
				Email:     email,
				User:      user,
			},
		})
	}

	return
}

func (p *ProviderImpl) generateSendVerificationCodeTasks(user model.User, loginIDs []loginid.LoginID) (tasks []async.TaskSpec) {
	for _, loginID := range loginIDs {
		for _, keyConfig := range p.UserVerificationConfiguration.LoginIDKeys {
			if keyConfig.Key == loginID.Key {
				tasks = append(tasks, async.TaskSpec{
					Name: task.VerifyCodeSendTaskName,
					Param: task.VerifyCodeSendTaskParam{
						URLPrefix: p.URLPrefixProvider.Value(),
						LoginID:   loginID.Value,
						UserID:    user.ID,
					},
				})
			}
		}
	}
	return
}
