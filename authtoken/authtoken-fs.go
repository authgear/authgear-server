package authtoken

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FileStore implements TokenStore by saving users' Token under
// a directory specified by a string. Each access token is
// stored in a separate file.
type FileStore struct {
	address string
	expiry  int64
}

// NewFileStore creates a file token store.
//
// It panics when it fails to create the directory.
func NewFileStore(address string, expiry int64) *FileStore {
	store := FileStore{address, expiry}
	err := os.MkdirAll(address, 0755)
	if err != nil {
		panic("FileStore.init: " + err.Error())
	}
	return &store
}

// NewToken creates a new token for this token store.
func (f *FileStore) NewToken(appName string, userInfoID string) (Token, error) {
	var expireAt time.Time
	if f.expiry > 0 {
		expireAt = time.Now().Add(time.Duration(f.expiry) * time.Second)
	}
	return New(appName, userInfoID, expireAt), nil
}

// Get tries to read the specified access token from file and
// writes to the supplied Token.
//
// Get returns an NotFoundError if no such access token exists or
// such access token is expired. In the latter case the expired
// access token is still written onto the supplied Token.
func (f *FileStore) Get(accessToken string, token *Token) error {
	if err := validateToken(accessToken); err != nil {
		return &NotFoundError{accessToken, err}
	}

	tokenPath := filepath.Join(f.address, accessToken)

	file, err := os.Open(tokenPath)
	if err != nil {
		return &NotFoundError{accessToken, err}
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(token); err != nil {
		return &NotFoundError{accessToken, err}
	}

	if token.IsExpired() {
		os.Remove(tokenPath)
		return &NotFoundError{accessToken, fmt.Errorf("token expired at %v", token.ExpiredAt)}
	}

	return nil
}

// Put writes the specified token into a file and overwrites existing
// Token if any.
func (f *FileStore) Put(token *Token) error {
	if err := validateToken(token.AccessToken); err != nil {
		return &NotFoundError{token.AccessToken, err}
	}

	file, err := os.Create(filepath.Join(f.address, token.AccessToken))
	if err != nil {
		return &NotFoundError{token.AccessToken, err}
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(token); err != nil {
		return &NotFoundError{token.AccessToken, err}
	}

	return nil
}

// Delete removes the access token from the file store.
//
// Delete return an error if the token cannot removed. It is NOT
// not an error if the token does not exist at deletion time.
func (f *FileStore) Delete(accessToken string) error {
	if err := validateToken(accessToken); err != nil {
		return &NotFoundError{accessToken, err}
	}

	if err := os.Remove(filepath.Join(f.address, accessToken)); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
