package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/twinj/uuid"
)

// Token is an expiry access token associated to a UserInfo.
type Token struct {
	AccessToken string    `json:"accessToken"`
	ExpiredAt   time.Time `json:"expiredAt"`
	UserInfoID  string    `json:"userInfoID"`
}

// NewToken creates a new Token ready for use given a userInfoID and
// expiredAt date. If expiredAt is passed an empty Time, it
// will be set to 30 days from now.
func NewToken(userInfoID string, expiredAt time.Time) Token {
	if expiredAt.IsZero() {
		expiredAt = time.Now().Add(24 * 30 * time.Hour)
	}

	return Token{
		// NOTE(limouren): I am not sure if it is good to use UUID
		// as access token.
		AccessToken: uuid.NewV4().String(),
		ExpiredAt:   expiredAt,
		UserInfoID:  userInfoID,
	}
}

// IsExpired determines whether the Token has expired now or not.
func (t *Token) IsExpired() bool {
	return t.ExpiredAt.Before(time.Now())
}

// TokenNotFoundError is the error returned by Get if a TokenStore
// cannot find the requested token or the fetched token is expired.
type TokenNotFoundError struct {
	AccessToken string
	Err         error
}

func (e *TokenNotFoundError) Error() string {
	return fmt.Sprintf("get %v: %v", e.AccessToken, e.Err)
}

// TokenStore represents a persistent storage for Token.
type TokenStore interface {
	Get(accessToken string, token *Token) error
	Put(token *Token) error
}

// FileStore implements TokenStore by saving users' Token under
// a directory specified by a string. Each access token is
// stored in a separate file.
type FileStore string

// Init MkAllDirs the FileStore directory and return itself.
//
// It panics when it fails to create the directory.
func (f FileStore) Init() FileStore {
	err := os.MkdirAll(string(f), 0755)
	if err != nil {
		panic("FileStore.init: " + err.Error())
	}
	return f
}

// Get tries to read the specified access token from file and
// writes to the supplied Token.
//
// Get returns an TokenNotFoundError if no such access token exists or
// such access token is expired. In the latter case the expired
// access token is still written onto the supplied Token.
func (f FileStore) Get(accessToken string, token *Token) error {
	tokenPath := filepath.Join(string(f), accessToken)

	file, err := os.Open(tokenPath)
	if err != nil {
		return &TokenNotFoundError{accessToken, err}
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(token); err != nil {
		return &TokenNotFoundError{accessToken, err}
	}

	if token.IsExpired() {
		os.Remove(tokenPath)
		return &TokenNotFoundError{accessToken, fmt.Errorf("token expired at %v", token.ExpiredAt)}
	}

	return nil
}

// Put writes the specified token into a file and overwrites existing
// Token if any.
func (f FileStore) Put(token *Token) error {
	file, err := os.Create(filepath.Join(string(f), token.AccessToken))
	if err != nil {
		return &TokenNotFoundError{token.AccessToken, err}
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		return &TokenNotFoundError{token.AccessToken, err}
	}

	return nil
}
