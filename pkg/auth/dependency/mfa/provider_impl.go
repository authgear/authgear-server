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
	sender           Sender
}

func NewProvider(store Store, mfaConfiguration config.MFAConfiguration, timeProvider time.Provider, sender Sender) Provider {
	return &providerImpl{
		store:            store,
		mfaConfiguration: mfaConfiguration,
		timeProvider:     timeProvider,
		sender:           sender,
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
					var token string
					token, err = p.createBearerToken(userID, aa.ID, now)
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
	totp, err := p.store.GetTOTP(userID, id)
	if err == nil {
		return p.DeleteTOTPAuthenticator(totp)
	}

	oob, err := p.store.GetOOB(userID, id)
	if err == nil {
		return p.DeleteOOBAuthenticator(oob)
	}

	return nil
}

func (p *providerImpl) DeleteTOTPAuthenticator(a *TOTPAuthenticator) error {
	authenticators, err := p.store.ListAuthenticators(a.UserID)
	if err != nil {
		return err
	}

	err = p.store.DeleteBearerTokenByParentID(a.UserID, a.ID)
	if err != nil {
		return err
	}

	err = p.store.DeleteTOTP(a)
	if err != nil {
		return err
	}

	deletingLastActivated := IsDeletingLastActivatedAuthenticator(authenticators, *a)
	if deletingLastActivated {
		err = p.store.DeleteRecoveryCode(a.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *providerImpl) DeleteOOBAuthenticator(a *OOBAuthenticator) error {
	authenticators, err := p.store.ListAuthenticators(a.UserID)
	if err != nil {
		return err
	}

	// To OOB authenticator, we have to delete its dependencies first
	// 1. bearer token
	// 2. OOB code

	err = p.store.DeleteBearerTokenByParentID(a.UserID, a.ID)
	if err != nil {
		return err
	}

	err = p.store.DeleteOOBCodeByAuthenticator(a)
	if err != nil {
		return err
	}

	err = p.store.DeleteOOB(a)
	if err != nil {
		return err
	}

	deletingLastActivated := IsDeletingLastActivatedAuthenticator(authenticators, *a)
	if deletingLastActivated {
		err = p.store.DeleteRecoveryCode(a.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *providerImpl) CreateOOB(userID string, channel coreAuth.AuthenticatorOOBChannel, phone string, email string) (*OOBAuthenticator, error) {
	now := p.timeProvider.NowUTC()
	a := OOBAuthenticator{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      coreAuth.AuthenticatorTypeOOB,
		CreatedAt: now,
		Channel:   channel,
		Phone:     phone,
		Email:     email,
	}
	authenticators, err := p.store.ListAuthenticators(a.UserID)
	if err != nil {
		return nil, err
	}
	ok := CanAddAuthenticator(authenticators, a, p.mfaConfiguration)
	if !ok {
		return nil, skyerr.NewError(skyerr.BadRequest, "no more authenticator can be added")
	}
	err = p.store.CreateOOB(&a)
	if err != nil {
		return nil, err
	}
	err = p.TriggerOOB(userID, a.ID)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (p *providerImpl) TriggerOOB(userID string, id string) (err error) {
	// Resolve the OOBAuthenticator
	// If id is given, simply get it by ID
	// Otherwise, loop through activated authenticators and assert the only one.
	var a *OOBAuthenticator
	if id != "" {
		a, err = p.store.GetOOB(userID, id)
		if err != nil {
			return
		}
	} else {
		authenticators, err := p.store.ListAuthenticators(userID)
		if err != nil {
			return err
		}
		var oob []OOBAuthenticator
		for _, b := range authenticators {
			switch v := b.(type) {
			case OOBAuthenticator:
				vv := v
				oob = append(oob, vv)
			default:
				break
			}
		}
		if len(oob) != 1 {
			return skyerr.NewError(skyerr.BadRequest, "must specify authenticator ID")
		}
		a = &oob[0]
	}

	// Find the existing valid code.
	// If not found, create a new one.
	now := p.timeProvider.NowUTC()
	oobCodes, err := p.store.GetValidOOBCode(userID, now)
	if err != nil {
		return
	}
	var oobCode *OOBCode
	for _, c := range oobCodes {
		if c.AuthenticatorID == a.ID {
			cc := c
			oobCode = &cc
			break
		}
	}
	if oobCode == nil {
		code := GenerateRandomOOBCode()
		// Assert code is unique
		for _, c := range oobCodes {
			if c.Code == code {
				err = skyerr.NewError(skyerr.UnexpectedError, "cannot generate oob code")
				return
			}
		}
		oobCode = &OOBCode{
			ID:              uuid.New(),
			UserID:          userID,
			AuthenticatorID: a.ID,
			Code:            code,
			CreatedAt:       now,
			// TODO(mfa): Allow customizing OOB code expiry
			ExpireAt: now.Add(5 * gotime.Minute),
		}
		err = p.store.CreateOOBCode(oobCode)
		if err != nil {
			return err
		}
	}

	err = p.sender.Send(oobCode.Code, a.Phone, a.Email)
	if err != nil {
		return err
	}

	return nil
}

func (p *providerImpl) ActivateOOB(userID string, id string, code string) ([]string, error) {
	a, err := p.store.GetOOB(userID, id)
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
	oobCodes, err := p.store.GetValidOOBCode(userID, now)
	if err != nil {
		return nil, err
	}

	// Find the OOB code
	var oobCode *OOBCode
	for _, c := range oobCodes {
		isTargetAuthenticator := c.AuthenticatorID == id
		isCodeValid := subtle.ConstantTimeCompare([]byte(code), []byte(c.Code)) == 1
		if isTargetAuthenticator && isCodeValid {
			cc := c
			oobCode = &cc
		}
	}
	if oobCode == nil {
		return nil, skyerr.NewError(skyerr.BadRequest, "invalid code")
	}

	// Delete the code so that it cannot be reused.
	err = p.store.DeleteOOBCode(oobCode)
	if err != nil {
		return nil, err
	}

	a.Activated = true
	a.ActivatedAt = &now
	err = p.store.UpdateOOB(a)
	if err != nil {
		return nil, err
	}

	generateRecoveryCode := len(authenticators) <= 0
	if generateRecoveryCode {
		return p.GenerateRecoveryCode(userID)
	}

	return nil, nil
}

func (p *providerImpl) createBearerToken(userID string, parentID string, now gotime.Time) (string, error) {
	expireAt := now.Add(gotime.Duration(p.mfaConfiguration.BearerToken.ExpireInDays) * gotime.Hour * 24)
	token := GenerateRandomBearerToken()
	bt := BearerTokenAuthenticator{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      coreAuth.AuthenticatorTypeBearerToken,
		ParentID:  parentID,
		Token:     token,
		CreatedAt: now,
		ExpireAt:  expireAt,
	}
	err := p.store.CreateBearerToken(&bt)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (p *providerImpl) AuthenticateOOB(userID string, code string, generateBearerToken bool) (*OOBAuthenticator, string, error) {
	now := p.timeProvider.NowUTC()
	oobCodes, err := p.store.GetValidOOBCode(userID, now)
	if err != nil {
		return nil, "", err
	}

	// Find the OOB code
	var oobCode *OOBCode
	for _, c := range oobCodes {
		isCodeValid := subtle.ConstantTimeCompare([]byte(code), []byte(c.Code)) == 1
		if isCodeValid {
			cc := c
			oobCode = &cc
		}
	}
	if oobCode == nil {
		return nil, "", skyerr.NewError(skyerr.BadRequest, "invalid code")
	}

	// Delete the code so that it cannot be reused.
	err = p.store.DeleteOOBCode(oobCode)
	if err != nil {
		return nil, "", err
	}

	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, "", err
	}

	for _, iface := range authenticators {
		switch a := iface.(type) {
		case OOBAuthenticator:
			if a.ID == oobCode.AuthenticatorID {
				aa := a
				if generateBearerToken {
					var token string
					token, err = p.createBearerToken(userID, aa.ID, now)
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

	return nil, "", skyerr.NewError(skyerr.BadRequest, "invalid code")
}

var (
	_ Provider = &providerImpl{}
)
