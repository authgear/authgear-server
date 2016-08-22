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

package authtokentest

import (
	"errors"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/authtoken"
)

// SingleTokenStore is a token store for storing a single auth token for testing.
type SingleTokenStore struct {
	Token *authtoken.Token
}

func (s *SingleTokenStore) NewToken(appName string, userInfoID string) (authtoken.Token, error) {
	return authtoken.New(appName, userInfoID, time.Time{}), nil
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
