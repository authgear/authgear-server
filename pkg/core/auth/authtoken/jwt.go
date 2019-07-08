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
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/skygeario/skygear-server/pkg/core/uuid"
)

type authTokenClaims struct {
	jwt.StandardClaims
	PrincipalID string `json:"skygear:principal_id"`
}

// JWTStore implements TokenStore by encoding user information into
// the access token string. This store does not keep state.
type JWTStore struct {
	appName string
	secret  string
	expiry  int64
}

// NewJWTStore creates a JWT token store.
func NewJWTStore(appName string, secret string, expiry int64) *JWTStore {
	if secret == "" {
		panic("jwt store is not configured with a secret")
	}
	store := JWTStore{
		appName: appName,
		secret:  secret,
		expiry:  expiry,
	}
	return &store
}

// NewToken creates a new token for this token store.
func (r *JWTStore) NewToken(authInfoID string, principalID string) (Token, error) {
	claims := authTokenClaims{
		StandardClaims: jwt.StandardClaims{
			Id:       uuid.New(),
			IssuedAt: time.Now().Unix(),
			Issuer:   r.appName,
			Subject:  authInfoID,
		},
		PrincipalID: principalID,
	}

	if r.expiry > 0 {
		claims.ExpiresAt = time.Now().Unix() + r.expiry
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := jwtToken.SignedString([]byte(r.secret))
	if err != nil {
		return Token{}, err
	}

	token := Token{}
	r.setTokenFromClaims(claims, &token)
	token.AccessToken = signedString
	return token, nil
}

// Get decodes and verifies the access token for user information. It returns
// the access token containing information about the user.
func (r *JWTStore) Get(accessToken string, token *Token) error {
	claims := authTokenClaims{}
	jwtToken, err := jwt.ParseWithClaims(accessToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &NotFoundError{accessToken, errors.New("unexpected algorithm in token")}
		}
		return []byte(r.secret), nil
	})

	if err != nil {
		return &NotFoundError{accessToken, err}
	}

	if jwtToken.Valid {
		token.AccessToken = accessToken
		r.setTokenFromClaims(claims, token)
	} else {
		return &NotFoundError{accessToken, errors.New("invalid token")}
	}

	// The token is considered valid by the JWTStore. (i.e. the token
	// has a valid signature and the signature is verified with the secret.)
	//
	// In skygear-server, the `InjectAuthIfPresent` preprocessor is
	// responsible for checking if the token is still valid. A token
	// might become invalid if the user has changed the password after the
	// token is generated.
	return nil
}

func (r *JWTStore) setTokenFromClaims(claims authTokenClaims, token *Token) {
	if claims.ExpiresAt > 0 {
		token.ExpiredAt = time.Unix(claims.ExpiresAt, 0)
	} else {
		token.ExpiredAt = time.Time{}
	}
	if claims.IssuedAt > 0 {
		token.IssuedAt = time.Unix(claims.IssuedAt, 0)
	} else {
		token.IssuedAt = time.Time{}
	}
	token.AppName = claims.Issuer
	token.AuthInfoID = claims.Subject
	token.PrincipalID = claims.PrincipalID
}

// Put does nothing because the JWT token store does not store token.
func (r *JWTStore) Put(token *Token) error {
	return nil
}

// Delete does nothing because the JWT token store does not store token.
func (r *JWTStore) Delete(accessToken string) error {
	return nil
}

// this ensures that our structure conform to certain interfaces.
var (
	_ Store = &JWTStore{}
)
