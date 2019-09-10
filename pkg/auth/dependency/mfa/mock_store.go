package mfa

import (
	"fmt"

	coreAuth "github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type MockStore struct {
	TimeProvider time.Provider
	RecoveryCode map[string][]RecoveryCodeAuthenticator
	TOTP         map[string][]TOTPAuthenticator
	OOB          map[string][]OOBAuthenticator
	BearerToken  map[string][]BearerTokenAuthenticator
}

func NewMockStore(timeProvider time.Provider) Store {
	return &MockStore{
		TimeProvider: timeProvider,
		RecoveryCode: map[string][]RecoveryCodeAuthenticator{},
		TOTP:         map[string][]TOTPAuthenticator{},
		OOB:          map[string][]OOBAuthenticator{},
		BearerToken:  map[string][]BearerTokenAuthenticator{},
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

func (s *MockStore) ListAuthenticators(userID string) ([]interface{}, error) {
	totp := s.TOTP[userID]
	oob := s.OOB[userID]
	output := []interface{}{}
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

var (
	_ Store = &MockStore{}
)
