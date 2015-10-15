package authtoken

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oursky/skygear/uuid"
)

// Token is an expiry access token associated to a UserInfo.
type Token struct {
	AccessToken string    `json:"accessToken"`
	ExpiredAt   time.Time `json:"expiredAt"`
	AppName     string    `json:"appName"`
	UserInfoID  string    `json:"userInfoID"`
}

func (t Token) MarshalJSON() ([]byte, error) {
	return json.Marshal(&jsonToken{
		t.AccessToken,
		jsonStamp(t.ExpiredAt),
		t.AppName,
		t.UserInfoID,
	})
}

func (t *Token) UnmarshalJSON(data []byte) (err error) {
	token := jsonToken{}
	if err := json.Unmarshal(data, &token); err != nil {
		return err
	}
	t.AccessToken = token.AccessToken
	t.ExpiredAt = time.Time(token.ExpiredAt)
	t.AppName = token.AppName
	t.UserInfoID = token.UserInfoID
	return nil
}

type jsonToken struct {
	AccessToken string    `json:"accessToken"`
	ExpiredAt   jsonStamp `json:"expiredAt"`
	AppName     string    `json:"appName"`
	UserInfoID  string    `json:"userInfoID"`
}

type jsonStamp time.Time

func (t jsonStamp) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	return json.Marshal(tt.UnixNano())
}

func (t *jsonStamp) UnmarshalJSON(data []byte) (err error) {
	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	*t = jsonStamp(time.Unix(0, i))
	return nil
}

// New creates a new Token ready for use given a userInfoID and
// expiredAt date. If expiredAt is passed an empty Time, it
// will be set to 30 days from now.
func New(appName string, userInfoID string, expiredAt time.Time) Token {
	if expiredAt.IsZero() {
		expiredAt = time.Now().Add(24 * 30 * time.Hour)
	}

	return Token{
		// NOTE(limouren): I am not sure if it is good to use UUID
		// as access token.
		AccessToken: uuid.New(),
		ExpiredAt:   expiredAt,
		AppName:     appName,
		UserInfoID:  userInfoID,
	}
}

// IsExpired determines whether the Token has expired now or not.
func (t *Token) IsExpired() bool {
	return t.ExpiredAt.Before(time.Now())
}

// NotFoundError is the error returned by Get if a TokenStore
// cannot find the requested token or the fetched token is expired.
type NotFoundError struct {
	AccessToken string
	Err         error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("get %v: %v", e.AccessToken, e.Err)
}

// Store represents a persistent storage for Token.
type Store interface {
	Get(accessToken string, token *Token) error
	Put(token *Token) error
	Delete(accessToken string) error
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
// Get returns an NotFoundError if no such access token exists or
// such access token is expired. In the latter case the expired
// access token is still written onto the supplied Token.
func (f FileStore) Get(accessToken string, token *Token) error {
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
	if err := os.Remove(filepath.Join(string(f), accessToken)); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
