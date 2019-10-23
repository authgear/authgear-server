package userverify

import (
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/errors"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal/password"
)

type Provider interface {
	CreateVerifyCode(principal *password.Principal) (*VerifyCode, error)
	VerifyUser(
		passwordProvider password.Provider,
		authStore authinfo.Store,
		authInfo *authinfo.AuthInfo,
		code string,
	) (*VerifyCode, error)
}

type providerImpl struct {
	codeGenerator CodeGenerator
	store         Store
	config        config.UserVerificationConfiguration
	time          time.Provider
}

func NewProvider(
	codeGenerator CodeGenerator,
	store Store,
	config config.UserVerificationConfiguration,
	time time.Provider,
) Provider {
	return &providerImpl{
		codeGenerator: codeGenerator,
		store:         store,
		config:        config,
		time:          time,
	}
}

func (provider *providerImpl) CreateVerifyCode(principal *password.Principal) (*VerifyCode, error) {
	_, isValid := provider.config.LoginIDKeys[principal.LoginIDKey]
	if !isValid {
		return nil, ErrUnknownLoginIDKey
	}

	code := provider.codeGenerator.Generate(principal.LoginIDKey)

	verifyCode := NewVerifyCode()
	verifyCode.UserID = principal.UserID
	verifyCode.LoginIDKey = principal.LoginIDKey
	verifyCode.LoginID = principal.LoginID
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = provider.time.NowUTC()

	if err := provider.store.CreateVerifyCode(&verifyCode); err != nil {
		return nil, errors.HandledWithMessage(err, "failed to create verification code")
	}

	return &verifyCode, nil
}

func (provider *providerImpl) VerifyUser(
	passwordProvider password.Provider,
	authStore authinfo.Store,
	authInfo *authinfo.AuthInfo,
	code string,
) (*VerifyCode, error) {
	verifyCode, err := provider.store.GetVerifyCodeByUser(authInfo.ID)
	if err != nil {
		if !errors.Is(err, ErrCodeNotFound) {
			err = NewUserVerificationFailed(InvalidCode, "invalid verification code")
		}
		return nil, err
	}

	if !verifyCode.Check(code) {
		return nil, NewUserVerificationFailed(InvalidCode, "invalid verification code")
	}

	if verifyCode.Consumed {
		return nil, NewUserVerificationFailed(UsedCode, "verification code is used")
	}

	principals, err := passwordProvider.GetPrincipalsByLoginID(
		verifyCode.LoginIDKey,
		verifyCode.LoginID,
	)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get principals to verify")
	}

	// filter principals belonging to the user
	userPrincipals := []*password.Principal{}
	for _, principal := range principals {
		if principal.UserID == authInfo.ID {
			userPrincipals = append(userPrincipals, principal)
		}
	}
	principals = userPrincipals

	if len(principals) == 0 {
		return nil, NewUserVerificationFailed(InvalidCode, "invalid verification code")
	}

	expiryTime := provider.config.LoginIDKeys[verifyCode.LoginIDKey].Expiry
	expireAt := verifyCode.CreatedAt.Add(gotime.Duration(expiryTime) * gotime.Second)
	if provider.time.NowUTC().After(expireAt) {
		return nil, NewUserVerificationFailed(InvalidCode, "verification code has expired")
	}

	err = provider.markUserVerified(passwordProvider, authStore, authInfo, verifyCode)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to mark user as verified")
	}

	return verifyCode, nil
}

func (provider *providerImpl) markUserVerified(
	passwordProvider password.Provider,
	authStore authinfo.Store,
	authInfo *authinfo.AuthInfo,
	verifyCode *VerifyCode,
) (err error) {
	if err = provider.store.MarkConsumed(verifyCode.ID); err != nil {
		return
	}

	principals, err := passwordProvider.GetPrincipalsByUserID(authInfo.ID)
	if err != nil {
		return
	}

	// Update user
	authInfo.VerifyInfo[verifyCode.LoginID] = true
	authInfo.Verified = IsUserVerified(
		authInfo.VerifyInfo,
		principals,
		provider.config.Criteria,
		provider.config.LoginIDKeys,
	)

	if err = authStore.UpdateAuth(authInfo); err != nil {
		return
	}

	return
}
