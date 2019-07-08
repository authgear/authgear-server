package authtoken

import (
	"errors"
	"time"
)

// MockStore is the memory implementation of auth token store
type MockStore struct {
	TokenMap map[string]Token
}

// NewMockStore create mock store with empty map
func NewMockStore() *MockStore {
	return &MockStore{
		TokenMap: map[string]Token{},
	}
}

func (s *MockStore) NewToken(authInfoID string, principalID string) (Token, error) {
	return New("mockAppName", authInfoID, principalID, time.Time{}), nil
}

func (s *MockStore) Get(accessToken string, token *Token) error {
	t, ok := s.TokenMap[accessToken]
	if !ok {
		return &NotFoundError{token.AccessToken, errors.New("not found")}
	}
	token = &t
	return nil
}

func (s *MockStore) Put(token *Token) error {
	s.TokenMap[token.AccessToken] = *token
	return nil
}

func (s *MockStore) Delete(accessToken string) error {
	_, ok := s.TokenMap[accessToken]
	if ok {
		delete(s.TokenMap, accessToken)
	}

	return nil
}

func (s *MockStore) GetTokensByAuthInfoID(authInfoID string) []Token {
	tokens := []Token{}
	for _, token := range s.TokenMap {
		if token.AuthInfoID == authInfoID {
			tokens = append(tokens, token)
		}
	}
	return tokens
}
