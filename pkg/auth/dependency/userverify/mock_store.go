package userverify

import (
	"errors"
	"time"
)

type MockStore struct {
	CodeByID map[string]VerifyCode
	Expiry   int64
}

func (m *MockStore) CreateVerifyCode(code *VerifyCode) error {
	m.CodeByID[code.ID] = *code
	m.updateCodeExpiry(code)
	return nil
}

func (m *MockStore) UpdateVerifyCode(code *VerifyCode) error {
	_, found := m.CodeByID[code.ID]
	if !found {
		return errors.New("code not found")
	}

	m.CodeByID[code.ID] = *code
	m.updateCodeExpiry(code)

	return nil
}

func (m *MockStore) GetVerifyCodeByCode(code string, vCode *VerifyCode) error {
	for _, c := range m.CodeByID {
		if c.Code == code {
			vCode.ID = c.ID
			vCode.UserID = c.UserID
			vCode.RecordKey = c.RecordKey
			vCode.RecordValue = c.RecordValue
			vCode.Code = c.Code
			vCode.Consumed = c.Consumed
			vCode.CreatedAt = c.CreatedAt
			m.updateCodeExpiry(vCode)
			return nil
		}
	}

	return errors.New("code not found")
}

func (m *MockStore) updateCodeExpiry(code *VerifyCode) {
	if m.Expiry > 0 {
		expireAt := code.CreatedAt.Add(time.Second * time.Duration(m.Expiry))
		code.expireAt = &expireAt
	} else {
		code.expireAt = nil
	}
}
