package mfa

import (
	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type providerImpl struct {
	store            Store
	mfaConfiguration config.MFAConfiguration
	timeProvider     time.Provider
}

func NewProvider(store Store, mfaConfiguration config.MFAConfiguration, timeProvider time.Provider) Provider {
	return &providerImpl{
		store:            store,
		mfaConfiguration: mfaConfiguration,
		timeProvider:     timeProvider,
	}
}

func (p *providerImpl) GetRecoveryCode(userID string) ([]string, error) {
	aa, err := p.store.GetRecoveryCode(userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(aa))
	for i, a := range aa {
		codes[i] = a.Code
	}
	return codes, nil
}

func (p *providerImpl) GenerateRecoveryCode(userID string) ([]string, error) {
	aa, err := p.store.GenerateRecoveryCode(userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(aa))
	for i, a := range aa {
		codes[i] = a.Code
	}
	return codes, nil
}

func (p *providerImpl) ListAuthenticators(userID string) ([]interface{}, error) {
	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, err
	}
	return MaskAuthenticators(authenticators), nil
}

func (p *providerImpl) CreateTOTP(userID string, displayName string) (*TOTPAuthenticator, error) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}
	now := p.timeProvider.NowUTC()
	a := TOTPAuthenticator{
		ID:          uuid.New(),
		UserID:      userID,
		Type:        coreAuth.AuthenticatorTypeTOTP,
		CreatedAt:   now,
		Secret:      secret,
		DisplayName: displayName,
	}
	authenticators, err := p.store.ListAuthenticators(a.UserID)
	if err != nil {
		return nil, err
	}
	ok := CanAddAuthenticator(authenticators, a, p.mfaConfiguration)
	if !ok {
		return nil, skyerr.NewError(skyerr.BadRequest, "no more authenticator can be added")
	}
	err = p.store.CreateTOTP(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (p *providerImpl) ActivateTOTP(userID string, id string, code string) ([]string, error) {
	a, err := p.store.GetTOTP(userID, id)
	if err != nil {
		return nil, err
	}
	if a.Activated {
		return nil, nil
	}

	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, err
	}

	ok := CanAddAuthenticator(authenticators, *a, p.mfaConfiguration)
	if !ok {
		return nil, skyerr.NewError(skyerr.BadRequest, "no more authenticator can be added")
	}

	now := p.timeProvider.NowUTC()
	ok, err = ValidateTOTP(a.Secret, code, now)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, skyerr.NewError(skyerr.BadRequest, "invalid OTP")
	}

	a.Activated = true
	a.ActivatedAt = &now
	err = p.store.UpdateTOTP(a)
	if err != nil {
		return nil, err
	}

	generateRecoveryCode := len(authenticators) <= 0
	if generateRecoveryCode {
		return p.GenerateRecoveryCode(userID)
	}

	return nil, nil
}

var (
	_ Provider = &providerImpl{}
)
