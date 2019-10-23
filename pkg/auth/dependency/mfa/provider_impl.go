package mfa

import (
	"crypto/subtle"
	gotime "time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/time"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

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
		return nil, errors.HandledWithMessage(err, "failed to get recovery code")
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
		return nil, errors.HandledWithMessage(err, "failed to generate recovery code")
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
		return nil, errors.HandledWithMessage(err, "failed to get recovery code")
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
			return nil, errors.HandledWithMessage(err, "failed to consume recovery code")
		}

		return &a, nil
	}

	return nil, errInvalidRecoveryCode
}

func (p *providerImpl) DeleteAllBearerToken(userID string) error {
	err := p.store.DeleteAllBearerToken(userID)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to delete all bearer tokens")
	}
	return nil
}

func (p *providerImpl) DeleteExpiredBearerToken(userID string) error {
	err := p.store.DeleteExpiredBearerToken(userID)
	if err != nil {
		return errors.HandledWithMessage(err, "failed to delete expired bearer tokens")
	}
	return nil
}

func (p *providerImpl) AuthenticateBearerToken(userID string, token string) (*BearerTokenAuthenticator, error) {
	a, err := p.store.GetBearerTokenByToken(userID, token)
	if err != nil {
		if errors.Is(err, ErrNoAuthenticators) {
			err = errInvalidBearerToken
		}
		return nil, err
	}
	now := p.timeProvider.NowUTC()
	if now.After(a.ExpireAt) {
		return nil, errInvalidBearerToken
	}
	return a, nil
}

func (p *providerImpl) ListAuthenticators(userID string) ([]Authenticator, error) {
	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list expired bearer tokens")
	}
	return MaskAuthenticators(authenticators), nil
}

func (p *providerImpl) CreateTOTP(userID string, displayName string) (*TOTPAuthenticator, error) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to generate TOTP secret")
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
		return nil, errors.HandledWithMessage(err, "failed to list TOTP")
	}
	ok := CanAddAuthenticator(authenticators, a, p.mfaConfiguration)
	if !ok {
		return nil, NewInvalidMFARequest(TooManyAuthenticator, "no more authenticator can be added")
	}
	err = p.store.DeleteInactiveTOTP(userID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to delete inactive TOTP")
	}
	err = p.store.CreateTOTP(&a)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to create TOTP")
	}
	return &a, nil
}

func (p *providerImpl) ActivateTOTP(userID string, code string) ([]string, error) {
	a, err := p.store.GetOnlyInactiveTOTP(userID)
	if err != nil {
		if errors.Is(err, ErrNoAuthenticators) {
			err = errAuthenticatorNotFound
		} else {
			err = errors.HandledWithMessage(err, "failed to get TOTP")
		}
		return nil, err
	}
	if a.Activated {
		return nil, nil
	}

	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list TOTP")
	}

	ok := CanAddAuthenticator(authenticators, *a, p.mfaConfiguration)
	if !ok {
		return nil, NewInvalidMFARequest(TooManyAuthenticator, "no more authenticator can be added")
	}

	now := p.timeProvider.NowUTC()
	ok = ValidateTOTP(a.Secret, code, now)
	if !ok {
		return nil, NewInvalidMFARequest(IncorrectCode, "incorrect TOTP code")
	}

	a.Activated = true
	a.ActivatedAt = &now
	err = p.store.UpdateTOTP(a)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to update TOTP")
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
		return nil, "", errors.HandledWithMessage(err, "failed to list TOTP")
	}

	now := p.timeProvider.NowUTC()
	for _, iface := range authenticators {
		switch a := iface.(type) {
		case TOTPAuthenticator:
			ok := ValidateTOTP(a.Secret, code, now)
			if ok {
				aa := a
				if generateBearerToken {
					var token string
					token, err = p.createBearerToken(userID, aa.ID, now)
					if err != nil {
						return nil, "", errors.HandledWithMessage(err, "failed to create bearer token")
					}
					return &aa, token, nil
				}
				return &aa, "", nil
			}
		default:
			break
		}
	}

	return nil, "", errInvalidMFACode
}

func (p *providerImpl) DeleteAuthenticator(userID string, id string) error {
	totp, err := p.store.GetTOTP(userID, id)
	if err == nil {
		err = p.deleteTOTPAuthenticator(totp)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to delete TOTP")
		}
		return err
	} else if !errors.Is(err, ErrNoAuthenticators) {
		return errors.HandledWithMessage(err, "failed to get TOTP")
	}

	oob, err := p.store.GetOOB(userID, id)
	if err == nil {
		err = p.deleteOOBAuthenticator(oob)
		if err != nil {
			err = errors.HandledWithMessage(err, "failed to delete OOB")
		}
		return err
	} else if !errors.Is(err, ErrNoAuthenticators) {
		return errors.HandledWithMessage(err, "failed to get OOB")
	}

	return errAuthenticatorNotFound
}

func (p *providerImpl) deleteTOTPAuthenticator(a *TOTPAuthenticator) error {
	authenticators, err := p.store.ListAuthenticators(a.UserID)
	if err != nil {
		return err
	}

	err = p.store.DeleteTOTP(a)
	if err != nil {
		return err
	}

	deletingLastActivated := IsDeletingOnlyActivatedAuthenticator(authenticators, *a)
	if deletingLastActivated {
		err = p.store.DeleteRecoveryCode(a.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *providerImpl) deleteOOBAuthenticator(a *OOBAuthenticator) error {
	authenticators, err := p.store.ListAuthenticators(a.UserID)
	if err != nil {
		return err
	}

	err = p.store.DeleteOOB(a)
	if err != nil {
		return err
	}

	deletingLastActivated := IsDeletingOnlyActivatedAuthenticator(authenticators, *a)
	if deletingLastActivated {
		err = p.store.DeleteRecoveryCode(a.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *providerImpl) CreateOOB(userID string, channel coreAuth.AuthenticatorOOBChannel, phone string, email string) (*OOBAuthenticator, error) {
	exceptID := ""
	createNew := false

	a, err := p.store.GetOOBByChannel(userID, channel, phone, email)
	// Forward any other errors.
	if err != nil && !errors.Is(err, ErrNoAuthenticators) {
		return nil, errors.HandledWithMessage(err, "failed to get OOB")
	}
	// Detect duplicate
	if err == nil && a.Activated {
		return nil, errAuthenticatorAlreadyExists
	}

	// If err is non-nil here, it must be ErrNoAuthenticators.
	if err != nil {
		createNew = true
		now := p.timeProvider.NowUTC()
		a = &OOBAuthenticator{
			ID:        uuid.New(),
			UserID:    userID,
			Type:      coreAuth.AuthenticatorTypeOOB,
			CreatedAt: now,
			Channel:   channel,
			Phone:     phone,
			Email:     email,
		}
	}
	exceptID = a.ID

	authenticators, err := p.store.ListAuthenticators(a.UserID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list OOB")
	}
	ok := CanAddAuthenticator(authenticators, *a, p.mfaConfiguration)
	if !ok {
		return nil, NewInvalidMFARequest(TooManyAuthenticator, "no more authenticator can be added")
	}

	err = p.store.DeleteInactiveOOB(userID, exceptID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to delete inactive OOB")
	}

	if createNew {
		err = p.store.CreateOOB(a)
		if err != nil {
			return nil, errors.HandledWithMessage(err, "failed to create OOB")
		}
	}

	err = p.TriggerOOB(userID, a.ID)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (p *providerImpl) TriggerOOB(userID string, id string) (err error) {
	// Resolve the OOBAuthenticator
	// If id is given, simply get it by ID
	// Otherwise, loop through activated authenticators and assert the only one.
	var a *OOBAuthenticator
	if id != "" {
		a, err = p.store.GetOOB(userID, id)
		if err != nil {
			if !errors.Is(err, ErrNoAuthenticators) {
				err = errors.HandledWithMessage(err, "failed to get OOB")
			}
			return err
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
			return NewInvalidMFARequest(AuthenticatorRequired, "must specify authenticator ID")
		}
		a = &oob[0]
	}

	// Find the existing valid code.
	// If not found, create a new one.
	now := p.timeProvider.NowUTC()
	oobCodes, err := p.store.GetValidOOBCode(userID, now)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to get OOB code")
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
				err = errors.New("generated OOB already exists")
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
			err = errors.HandledWithMessage(err, "failed to create OOB code")
			return err
		}
	}

	// TODO(mfa): Do not always send OOB code.
	err = p.sender.Send(oobCode.Code, a.Phone, a.Email)
	if err != nil {
		err = errors.HandledWithMessage(err, "failed to send OOB code")
		return err
	}

	return nil
}

func (p *providerImpl) ActivateOOB(userID string, code string) ([]string, error) {
	a, err := p.store.GetOnlyInactiveOOB(userID)
	if err != nil {
		if errors.Is(err, ErrNoAuthenticators) {
			err = errAuthenticatorNotFound
		} else {
			err = errors.HandledWithMessage(err, "failed to get OOB")
		}
		return nil, err
	}
	if a.Activated {
		return nil, nil
	}

	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to list OOB")
	}

	ok := CanAddAuthenticator(authenticators, *a, p.mfaConfiguration)
	if !ok {
		return nil, NewInvalidMFARequest(TooManyAuthenticator, "no more authenticator can be added")
	}

	now := p.timeProvider.NowUTC()
	oobCodes, err := p.store.GetValidOOBCode(userID, now)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to get OOB code")
	}

	// Find the OOB code
	var oobCode *OOBCode
	for _, c := range oobCodes {
		isTargetAuthenticator := c.AuthenticatorID == a.ID
		isCodeValid := subtle.ConstantTimeCompare([]byte(code), []byte(c.Code)) == 1
		if isTargetAuthenticator && isCodeValid {
			cc := c
			oobCode = &cc
		}
	}
	if oobCode == nil {
		return nil, NewInvalidMFARequest(IncorrectCode, "incorrect OOB code")
	}

	// Delete the code so that it cannot be reused.
	err = p.store.DeleteOOBCode(oobCode)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to delete OOB code")
	}

	a.Activated = true
	a.ActivatedAt = &now
	err = p.store.UpdateOOB(a)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "failed to update OOB")
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
		return nil, "", errors.HandledWithMessage(err, "failed to get OOB code")
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
		return nil, "", errInvalidMFACode
	}

	// Delete the code so that it cannot be reused.
	err = p.store.DeleteOOBCode(oobCode)
	if err != nil {
		return nil, "", errors.HandledWithMessage(err, "failed to delete OOB code")
	}

	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, "", errors.HandledWithMessage(err, "failed to list OOB")
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
						return nil, "", errors.HandledWithMessage(err, "failed to create bearer token")
					}
					return &aa, token, nil
				}
				return &aa, "", nil
			}
		default:
			break
		}
	}

	return nil, "", errInvalidMFACode
}

func (p *providerImpl) StepMFA(a *coreAuth.AuthnSession, opts coreAuth.AuthnSessionStepMFAOptions) error {
	now := p.timeProvider.NowUTC()
	step, ok := a.NextStep()
	if !ok || step != coreAuth.AuthnSessionStepMFA {
		return skyerr.NewBadRequest("expected step to be mfa")
	}
	a.AuthenticatorID = opts.AuthenticatorID
	a.AuthenticatorType = opts.AuthenticatorType
	a.AuthenticatorOOBChannel = opts.AuthenticatorOOBChannel
	a.AuthenticatorUpdatedAt = &now
	a.AuthenticatorBearerToken = opts.AuthenticatorBearerToken
	a.FinishedSteps = append(a.FinishedSteps, step)
	return nil
}

var (
	_ Provider = &providerImpl{}
)
