// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authtoken

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/skygeario/skygear-server/uuid"
)

// Token is an expiry access token associated to a UserInfo.
type Token struct {
	AccessToken string    `json:"accessToken" redis:"accessToken"`
	ExpiredAt   time.Time `json:"expiredAt" redis:"expiredAt"`
	AppName     string    `json:"appName" redis:"appName"`
	UserInfoID  string    `json:"userInfoID" redis:"userInfoID"`
}

// MarshalJSON implements the json.Marshaler interface.
func (t Token) MarshalJSON() ([]byte, error) {
	var expireAt jsonStamp
	if !t.ExpiredAt.IsZero() {
		expireAt = jsonStamp(t.ExpiredAt)
	}
	return json.Marshal(&jsonToken{
		t.AccessToken,
		expireAt,
		t.AppName,
		t.UserInfoID,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Token) UnmarshalJSON(data []byte) (err error) {
	token := jsonToken{}
	if err := json.Unmarshal(data, &token); err != nil {
		return err
	}
	var expireAt time.Time
	if !time.Time(token.ExpiredAt).IsZero() {
		expireAt = time.Time(token.ExpiredAt)
	}
	t.AccessToken = token.AccessToken
	t.ExpiredAt = expireAt
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

// MarshalJSON implements the json.Marshaler interface.
func (t jsonStamp) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return json.Marshal(0)
	}
	return json.Marshal(tt.UnixNano())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *jsonStamp) UnmarshalJSON(data []byte) (err error) {
	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	if i == 0 {
		*t = jsonStamp{}
		return nil
	}
	*t = jsonStamp(time.Unix(0, i))
	return nil
}

// New creates a new Token ready for use given a userInfoID and
// expiredAt date. If expiredAt is passed an empty Time, the token
// does not expire.
func New(appName string, userInfoID string, expiredAt time.Time) Token {
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
	return !t.ExpiredAt.IsZero() && t.ExpiredAt.Before(time.Now())
}

// NotFoundError is the error returned by Get if a TokenStore
// cannot find the requested token or the fetched token is expired.
type NotFoundError struct {
	AccessToken string
	Err         error
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("get %#v: %v", e.AccessToken, e.Err)
}

// Store represents a persistent storage for Token.
type Store interface {
	Get(accessToken string, token *Token) error
	Put(token *Token) error
	Delete(accessToken string) error
}

var errInvalidToken = errors.New("invalid access token")

func validateToken(base string) error {
	b := filepath.Base(base)
	if b != base || b == "." || b == "/" {
		return errInvalidToken
	}
	return nil
}

// Configuration encapsulates arguments to initialize a token store
type Configuration struct {
	Implementation string
	Path           string
	Prefix         string
}

// InitTokenStore accept a implementation and path string. Return a Store.
func InitTokenStore(config Configuration) Store {
	var store Store
	switch config.Implementation {
	default:
		panic("unrecgonized token store implementation: " + config.Implementation)
	case "fs":
		store = NewFileStore(config.Path)
	case "redis":
		store = NewRedisStore(config.Path, config.Prefix)
	}
	return store
}
