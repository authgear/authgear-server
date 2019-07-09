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
	"fmt"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

// Token is an expiry access token associated to a AuthInfo.
type Token struct {
	AccessToken string    `json:"accessToken" redis:"accessToken"`
	ExpiredAt   time.Time `json:"expiredAt" redis:"expiredAt"`
	AppName     string    `json:"appName" redis:"appName"`
	AuthInfoID  string    `json:"authInfoID" redis:"authInfoID"`
	IssuedAt    time.Time `json:"issuedAt" redis:"issuedAt"`
}

// MarshalJSON implements the json.Marshaler interface.
func (t Token) MarshalJSON() ([]byte, error) {
	var expireAt, issuedAt jsonStamp
	if !t.ExpiredAt.IsZero() {
		expireAt = jsonStamp(t.ExpiredAt)
	}
	if !t.IssuedAt.IsZero() {
		issuedAt = jsonStamp(t.IssuedAt)
	}
	return json.Marshal(&jsonToken{
		t.AccessToken,
		expireAt,
		t.AppName,
		t.AuthInfoID,
		issuedAt,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Token) UnmarshalJSON(data []byte) (err error) {
	token := jsonToken{}
	if err := json.Unmarshal(data, &token); err != nil {
		return err
	}
	var expireAt, issuedAt time.Time
	if !time.Time(token.ExpiredAt).IsZero() {
		expireAt = time.Time(token.ExpiredAt)
	}
	if !time.Time(token.IssuedAt).IsZero() {
		issuedAt = time.Time(token.IssuedAt)
	}
	t.AccessToken = token.AccessToken
	t.ExpiredAt = expireAt
	t.AppName = token.AppName
	t.AuthInfoID = token.AuthInfoID
	t.IssuedAt = issuedAt
	return nil
}

type jsonToken struct {
	AccessToken string    `json:"accessToken"`
	ExpiredAt   jsonStamp `json:"expiredAt"`
	AppName     string    `json:"appName"`
	AuthInfoID  string    `json:"authInfoID"`
	IssuedAt    jsonStamp `json:"issuedAt"`
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

// New creates a new Token ready for use given a authInfoID and
// expiredAt date. If expiredAt is passed an empty Time, the token
// does not expire.
func New(appName string, authInfoID string, expiredAt time.Time) Token {
	return Token{
		// NOTE(limouren): I am not sure if it is good to use UUID
		// as access token.
		AccessToken: uuid.New(),
		ExpiredAt:   expiredAt,
		AppName:     appName,
		AuthInfoID:  authInfoID,
		IssuedAt:    time.Now().UTC(),
	}
}

// IsExpired determines whether the Token has expired now or not.
func (t *Token) IsExpired() bool {
	return !t.ExpiredAt.IsZero() && t.ExpiredAt.Before(time.Now().UTC())
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
