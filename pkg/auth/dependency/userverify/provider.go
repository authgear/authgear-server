package userverify

import (
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/errors"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/identity/loginid"
)

type LoginIDProvider interface {
	GetByLoginID(loginid.LoginID) ([]*loginid.Identity, error)
	List(userID string) ([]*loginid.Identity, error)
}

type Provider interface {
	CreateVerifyCode(*loginid.Identity) (*VerifyCode, error)
	VerifyUser(
		loginIDProvider LoginIDProvider,
		authStore authinfo.Store,
		authInfo *authinfo.AuthInfo,
		code string,
	) (*VerifyCode, error)
	UpdateVerificationState(
		authInfo *authinfo.AuthInfo,
		authStore authinfo.Store,
		identities []*loginid.Identity,
	) error
}

type providerImpl struct {
	codeGenerator CodeGenerator
	store         Store
	config        *config.UserVerificationConfiguration
	time          time.Provider
}

func NewProvider(
	codeGenerator CodeGenerator,
	store Store,
	config *config.UserVerificationConfiguration,
	time time.Provider,
) Provider {
	return &providerImpl{
		codeGenerator: codeGenerator,
		store:         store,
		config:        config,
		time:          time,
	}
}

func (provider *providerImpl) CreateVerifyCode(i *loginid.Identity) (*VerifyCode, error) {
	_, isValid := provider.config.GetLoginIDKey(i.LoginIDKey)
	if !isValid {
		return nil, ErrUnknownLoginIDKey
	}

	code := provider.codeGenerator.Generate(i.LoginIDKey)

	verifyCode := NewVerifyCode()
	verifyCode.UserID = i.UserID
	verifyCode.LoginIDKey = i.LoginIDKey
	verifyCode.LoginID = i.LoginID
	verifyCode.Code = code
	verifyCode.Consumed = false
	verifyCode.CreatedAt = provider.time.NowUTC()

	if err := provider.store.CreateVerifyCode(&verifyCode); err != nil {
		return nil, errors.HandledWithMessage(err, "failed to create verification code")
	}

	return &verifyCode, nil
}

func (provider *providerImpl) VerifyUser(
	loginIDProvider LoginIDProvider,
	authStore authinfo.Store,
	authInfo *authinfo.AuthInfo,
	code string,
) (*VerifyCode, error) {
	verifyCode, err := provider.store.GetVerifyCodeByUser(authInfo.ID)
	if err != nil {
		if errors.Is(err, ErrCodeNotFound) {
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

	is, err := loginIDProvider.GetByLoginID(loginid.LoginID{
		Key:   verifyCode.LoginIDKey,
		Value: verifyCode.LoginID,
	})
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get identities to verify")
	}

	// filter principals belonging to the user
	var identities []*loginid.Identity
	for _, i := range is {
		if i.UserID == authInfo.ID {
			identities = append(identities, i)
		}
	}

	if len(identities) == 0 {
		return nil, NewUserVerificationFailed(InvalidCode, "invalid verification code")
	}

	c, ok := provider.config.GetLoginIDKey(verifyCode.LoginIDKey)
	if !ok {
		panic("invalid login id key: " + verifyCode.LoginIDKey)
	}
	expiryTime := c.Expiry
	expireAt := verifyCode.CreatedAt.Add(gotime.Duration(expiryTime) * gotime.Second)
	if provider.time.NowUTC().After(expireAt) {
		return nil, NewUserVerificationFailed(ExpiredCode, "verification code has expired")
	}

	err = provider.markUserVerified(loginIDProvider, authStore, authInfo, verifyCode)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to mark user as verified")
	}

	return verifyCode, nil
}

func (provider *providerImpl) markUserVerified(
	loginIDProvider LoginIDProvider,
	authStore authinfo.Store,
	authInfo *authinfo.AuthInfo,
	verifyCode *VerifyCode,
) (err error) {
	if err = provider.store.MarkConsumed(verifyCode.ID); err != nil {
		return
	}

	is, err := loginIDProvider.List(authInfo.ID)
	if err != nil {
		return
	}

	// Update user
	authInfo.VerifyInfo[verifyCode.LoginID] = true
	if err = provider.UpdateVerificationState(authInfo, authStore, is); err != nil {
		return
	}

	return
}

func (provider *providerImpl) UpdateVerificationState(
	authInfo *authinfo.AuthInfo,
	authStore authinfo.Store,
	identities []*loginid.Identity,
) error {
	isVerified := IsUserVerified(
		authInfo.VerifyInfo,
		identities,
		provider.config.Criteria,
		provider.config.LoginIDKeys,
	)

	authInfo.Verified = isVerified
	if err := authStore.UpdateAuth(authInfo); err != nil {
		return err
	}

	return nil
}
