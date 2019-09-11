package mfa

import (
	"crypto/subtle"
	"errors"
	gotime "time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

var ErrInvalidRecoveryCode = errors.New("invalid recovery code")
var ErrInvalidBearerToken = errors.New("invalid bearer token")

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

func (p *providerImpl) AuthenticateRecoveryCode(userID string, code string) (*RecoveryCodeAuthenticator, error) {
	recoveryCodes, err := p.store.GetRecoveryCode(userID)
	if err != nil {
		return nil, err
	}

	for _, recoveryCode := range recoveryCodes {
		if recoveryCode.Consumed {
			continue
		}
		eq := subtle.ConstantTimeCompare([]byte(code), []byte(recoveryCode.Code)) == 1
		if !eq {
			continue
		}

		a := recoveryCode
		a.Consumed = true

		err = p.store.UpdateRecoveryCode(&a)
		if err != nil {
			return nil, err
		}

		return &a, nil
	}

	return nil, ErrInvalidRecoveryCode
}

func (p *providerImpl) DeleteAllBearerToken(userID string) error {
	return p.store.DeleteAllBearerToken(userID)
}

func (p *providerImpl) AuthenticateBearerToken(userID string, token string) (*BearerTokenAuthenticator, error) {
	a, err := p.store.GetBearerTokenByToken(userID, token)
	if err != nil {
		if err == ErrAuthenticatorNotFound {
			err = ErrInvalidBearerToken
		}
		return nil, err
	}
	now := p.timeProvider.NowUTC()
	if now.After(a.ExpireAt) {
		return nil, ErrInvalidBearerToken
	}
	return a, nil
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

func (p *providerImpl) AuthenticateTOTP(userID string, code string, generateBearerToken bool) (*TOTPAuthenticator, string, error) {
	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, "", err
	}

	now := p.timeProvider.NowUTC()
	for _, iface := range authenticators {
		switch a := iface.(type) {
		case TOTPAuthenticator:
			ok, err := ValidateTOTP(a.Secret, code, now)
			if err != nil {
				return nil, "", err
			}
			if ok {
				aa := a
				if generateBearerToken {
					expireAt := now.Add(gotime.Duration(p.mfaConfiguration.BearerToken.ExpireInDays) * gotime.Hour * 24)
					token := GenerateRandomBearerToken()
					bt := BearerTokenAuthenticator{
						ID:        uuid.New(),
						UserID:    userID,
						Type:      coreAuth.AuthenticatorTypeBearerToken,
						ParentID:  aa.ID,
						Token:     token,
						CreatedAt: now,
						ExpireAt:  expireAt,
					}
					err = p.store.CreateBearerToken(&bt)
					if err != nil {
						return nil, "", err
					}
					return &aa, token, nil
				}
				return &aa, "", nil
			}
		default:
			break
		}
	}

	return nil, "", skyerr.NewError(skyerr.BadRequest, "invalid OTP")
}

func (p *providerImpl) DeleteAuthenticator(userID string, id string) error {
	// TODO: Delete OOB
	a, err := p.store.GetTOTP(userID, id)
	if err != nil {
		return err
	}

	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return err
	}

	err = p.store.DeleteBearerTokenByParentID(userID, a.ID)
	if err != nil {
		return err
	}

	err = p.store.DeleteTOTP(a)
	if err != nil {
		return err
	}

	deletingLastActivated := IsDeletingLastActivatedAuthenticator(authenticators, *a)
	if deletingLastActivated {
		err = p.store.DeleteRecoveryCode(userID)
		if err != nil {
			return err
		}
	}

	return nil
}

var (
	_ Provider = &providerImpl{}
)
