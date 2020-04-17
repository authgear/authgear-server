package sso

import (
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

func NewMockSkygearAuthorizationCodeStore() SkygearAuthorizationCodeStore {
	return &MockSkygearAuthorizationCodeStore{
		codeMap: map[string]*SkygearAuthorizationCode{},
	}
}

type MockSkygearAuthorizationCodeStore struct {
	codeMap map[string]*SkygearAuthorizationCode
}

var _ SkygearAuthorizationCodeStore = &MockSkygearAuthorizationCodeStore{}

func (m *MockSkygearAuthorizationCodeStore) Get(codeHash string) (*SkygearAuthorizationCode, error) {
	if code, ok := m.codeMap[codeHash]; ok {
		return code, nil
	}
	return nil, ErrCodeNotFound
}

func (m *MockSkygearAuthorizationCodeStore) Set(code *SkygearAuthorizationCode) (err error) {
	if _, ok := m.codeMap[code.CodeHash]; ok {
		return errors.New("duplicated authorization code")
	}
	m.codeMap[code.CodeHash] = code
	return nil
}

func (m *MockSkygearAuthorizationCodeStore) Delete(codeHash string) error {
	delete(m.codeMap, codeHash)
	return nil
}
