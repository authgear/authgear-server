package authtokentest

import (
	"errors"

	"github.com/oursky/skygear/authtoken"
)

// SingleTokenStore is a token store for storing a single auth token for testing.
type SingleTokenStore struct {
	Token *authtoken.Token
}

func (s *SingleTokenStore) Get(accessToken string, token *authtoken.Token) error {
	if s.Token == nil {
		return &authtoken.NotFoundError{token.AccessToken, errors.New("not found")}
	}
	*token = authtoken.Token(*s.Token)
	return nil
}

func (s *SingleTokenStore) Put(token *authtoken.Token) error {
	newToken := authtoken.Token(*token)
	s.Token = &newToken
	return nil
}

func (s *SingleTokenStore) Delete(accessToken string) error {
	s.Token = nil
	return nil
}
