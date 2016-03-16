package authtoken

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// FileStore implements TokenStore by saving users' Token under
// a directory specified by a string. Each access token is
// stored in a separate file.
type FileStore string

// NewFileStore creates a file token store.
//
// It panics when it fails to create the directory.
func NewFileStore(address string) *FileStore {
	store := FileStore(address)
	err := os.MkdirAll(address, 0755)
	if err != nil {
		panic("FileStore.init: " + err.Error())
	}
	return &store
}

// Get tries to read the specified access token from file and
// writes to the supplied Token.
//
// Get returns an NotFoundError if no such access token exists or
// such access token is expired. In the latter case the expired
// access token is still written onto the supplied Token.
func (f FileStore) Get(accessToken string, token *Token) error {
	if err := validateToken(accessToken); err != nil {
		return &NotFoundError{accessToken, err}
	}

	tokenPath := filepath.Join(string(f), accessToken)

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
func (f FileStore) Put(token *Token) error {
	if err := validateToken(token.AccessToken); err != nil {
		return &NotFoundError{token.AccessToken, err}
	}

	file, err := os.Create(filepath.Join(string(f), token.AccessToken))
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
func (f FileStore) Delete(accessToken string) error {
	if err := validateToken(accessToken); err != nil {
		return &NotFoundError{accessToken, err}
	}

	if err := os.Remove(filepath.Join(string(f), accessToken)); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
