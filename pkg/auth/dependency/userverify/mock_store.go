package userverify

import (
	"errors"
)

type MockStore struct {
	CodeByID []VerifyCode
}

func (m *MockStore) CreateVerifyCode(code *VerifyCode) error {
	m.CodeByID = append(m.CodeByID, *code)
	return nil
}

func (m *MockStore) MarkConsumed(codeID string) error {
	for i, code := range m.CodeByID {
		if code.ID != codeID {
			continue
		}
		code.Consumed = true
		m.CodeByID[i] = code
		return nil
	}
	return errors.New("code not found")
}

func (m *MockStore) GetVerifyCodeByUser(userID string) (*VerifyCode, error) {
	for _, code := range m.CodeByID {
		if code.UserID == userID {
			return &code, nil
		}
	}

	return nil, errors.New("code not found")
}

var _ Store = &MockStore{}
