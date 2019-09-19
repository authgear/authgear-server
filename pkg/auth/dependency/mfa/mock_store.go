package mfa

import (
	"fmt"
	gotime "time"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type MockStore struct {
	TimeProvider time.Provider
	RecoveryCode map[string][]RecoveryCodeAuthenticator
	TOTP         map[string][]TOTPAuthenticator
	OOB          map[string][]OOBAuthenticator
	BearerToken  map[string][]BearerTokenAuthenticator
	OOBCode      map[string][]OOBCode
}

func NewMockStore(timeProvider time.Provider) Store {
	return &MockStore{
		TimeProvider: timeProvider,
		RecoveryCode: map[string][]RecoveryCodeAuthenticator{},
		TOTP:         map[string][]TOTPAuthenticator{},
		OOB:          map[string][]OOBAuthenticator{},
		BearerToken:  map[string][]BearerTokenAuthenticator{},
		OOBCode:      map[string][]OOBCode{},
	}
}

func (s *MockStore) GetRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error) {
	return s.RecoveryCode[userID], nil
}

func (s *MockStore) GenerateRecoveryCode(userID string) ([]RecoveryCodeAuthenticator, error) {
	now := s.TimeProvider.NowUTC()
	count := 16
	recoveryCodes := make([]RecoveryCodeAuthenticator, count)
	for i := 0; i < count; i++ {
		recoveryCodes[i] = RecoveryCodeAuthenticator{
			ID:        fmt.Sprintf("recovery-code-user-%s-%d", userID, count),
			UserID:    userID,
			Type:      coreAuth.AuthenticatorTypeRecoveryCode,
			Code:      GenerateRandomRecoveryCode(),
			CreatedAt: now,
			Consumed:  false,
		}
	}
	s.RecoveryCode[userID] = recoveryCodes
	return recoveryCodes, nil
}

func (s *MockStore) DeleteRecoveryCode(userID string) error {
	delete(s.RecoveryCode, userID)
	return nil
}

func (s *MockStore) UpdateRecoveryCode(a *RecoveryCodeAuthenticator) error {
	recoveryCode := s.RecoveryCode[a.UserID]
	for i, b := range recoveryCode {
		if b.ID == a.ID {
			recoveryCode[i] = *a
			return nil
		}
	}
	return ErrAuthenticatorNotFound
}

func (s *MockStore) DeleteBearerTokenByParentID(userID string, parentID string) error {
	bt := s.BearerToken[userID]
	var newBT []BearerTokenAuthenticator
	for _, a := range bt {
		if a.ParentID == parentID {
			continue
		}
		newBT = append(newBT, a)
	}
	s.BearerToken[userID] = newBT
	return nil
}

func (s *MockStore) DeleteAllBearerToken(userID string) error {
	delete(s.BearerToken, userID)
	return nil
}

func (s *MockStore) CreateBearerToken(a *BearerTokenAuthenticator) error {
	bt := s.BearerToken[a.UserID]
	if bt == nil {
		bt = []BearerTokenAuthenticator{*a}
	} else {
		bt = append(bt, *a)
	}
	s.BearerToken[a.UserID] = bt
	return nil
}

func (s *MockStore) GetBearerTokenByToken(userID string, token string) (*BearerTokenAuthenticator, error) {
	bt := s.BearerToken[userID]
	for _, a := range bt {
		if a.Token == token {
			aa := a
			return &aa, nil
		}
	}
	return nil, ErrAuthenticatorNotFound
}

func (s *MockStore) ListAuthenticators(userID string) ([]Authenticator, error) {
	totp := s.TOTP[userID]
	oob := s.OOB[userID]
	output := []Authenticator{}
	for _, a := range totp {
		if a.Activated {
			output = append(output, a)
		}
	}
	for _, a := range oob {
		if a.Activated {
			output = append(output, a)
		}
	}
	return output, nil
}

func (s *MockStore) CreateTOTP(a *TOTPAuthenticator) error {
	totp := s.TOTP[a.UserID]
	if totp == nil {
		totp = []TOTPAuthenticator{*a}
	} else {
		totp = append(totp, *a)
	}
	s.TOTP[a.UserID] = totp
	return nil
}

func (s *MockStore) GetTOTP(userID string, id string) (*TOTPAuthenticator, error) {
	totp := s.TOTP[userID]
	for _, a := range totp {
		if a.ID == id {
			aa := a
			return &aa, nil
		}
	}
	return nil, ErrAuthenticatorNotFound
}

func (s *MockStore) UpdateTOTP(a *TOTPAuthenticator) error {
	totp := s.TOTP[a.UserID]
	for i, b := range totp {
		if b.ID == a.ID {
			totp[i] = *a
			return nil
		}
	}
	return ErrAuthenticatorNotFound
}

func (s *MockStore) DeleteTOTP(a *TOTPAuthenticator) error {
	totp := s.TOTP[a.UserID]
	var newTOTP []TOTPAuthenticator
	for _, b := range totp {
		if b.ID == a.ID {
			continue
		}
		newTOTP = append(newTOTP, b)
	}
	s.TOTP[a.UserID] = newTOTP
	return nil
}

func (s *MockStore) DeleteInactiveTOTP(userID string) error {
	totp := s.TOTP[userID]
	var newTOTP []TOTPAuthenticator
	for _, b := range totp {
		if !b.Activated {
			continue
		}
		newTOTP = append(newTOTP, b)
	}
	s.TOTP[userID] = newTOTP
	return nil
}

func (s *MockStore) GetOnlyInactiveTOTP(userID string) (*TOTPAuthenticator, error) {
	totp := s.TOTP[userID]
	var output []TOTPAuthenticator
	for _, b := range totp {
		if !b.Activated {
			output = append(output, b)
		}
	}
	if len(output) != 1 {
		return nil, ErrAuthenticatorNotFound
	}
	return &output[0], nil
}

func (s *MockStore) CreateOOB(a *OOBAuthenticator) error {
	oob := s.OOB[a.UserID]
	if oob == nil {
		oob = []OOBAuthenticator{*a}
	} else {
		oob = append(oob, *a)
	}
	s.OOB[a.UserID] = oob
	return nil
}

func (s *MockStore) GetOOB(userID string, id string) (*OOBAuthenticator, error) {
	oob := s.OOB[userID]
	for _, a := range oob {
		if a.ID == id {
			aa := a
			return &aa, nil
		}
	}
	return nil, ErrAuthenticatorNotFound
}

func (s *MockStore) UpdateOOB(a *OOBAuthenticator) error {
	oob := s.OOB[a.UserID]
	for i, b := range oob {
		if b.ID == a.ID {
			oob[i] = *a
			return nil
		}
	}
	return ErrAuthenticatorNotFound
}

func (s *MockStore) DeleteOOB(a *OOBAuthenticator) error {
	oob := s.OOB[a.UserID]
	var newOOB []OOBAuthenticator
	for _, b := range oob {
		if b.ID == a.ID {
			continue
		}
		newOOB = append(newOOB, b)
	}
	s.OOB[a.UserID] = newOOB
	return nil
}

func (s *MockStore) GetValidOOBCode(userID string, t gotime.Time) ([]OOBCode, error) {
	oobCode := s.OOBCode[userID]
	var output []OOBCode
	for _, code := range oobCode {
		if code.ExpireAt.After(t) {
			c := code
			output = append(output, c)
		}
	}
	return output, nil
}

func (s *MockStore) CreateOOBCode(c *OOBCode) error {
	oobCode := s.OOBCode[c.UserID]
	if oobCode == nil {
		oobCode = []OOBCode{*c}
	} else {
		oobCode = append(oobCode, *c)
	}
	s.OOBCode[c.UserID] = oobCode
	return nil
}

func (s *MockStore) DeleteOOBCode(c *OOBCode) error {
	oobCode := s.OOBCode[c.UserID]
	var newOOBCode []OOBCode
	for _, d := range oobCode {
		if d.ID == c.ID {
			continue
		}
		newOOBCode = append(newOOBCode, d)
	}
	s.OOBCode[c.UserID] = newOOBCode
	return nil
}

func (s *MockStore) DeleteOOBCodeByAuthenticator(a *OOBAuthenticator) error {
	oobCode := s.OOBCode[a.UserID]
	var newOOBCode []OOBCode
	for _, d := range oobCode {
		if d.AuthenticatorID == a.ID {
			continue
		}
		newOOBCode = append(newOOBCode, d)
	}
	s.OOBCode[a.UserID] = newOOBCode
	return nil
}

var (
	_ Store = &MockStore{}
)
