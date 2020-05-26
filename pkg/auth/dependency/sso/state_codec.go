package sso

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
)

type stateClaims struct {
	State
	jwt.StandardClaims
}

type StateCodec struct {
	AppID       string
	OAuthConfig *config.OAuthConfiguration
}

func NewStateCodec(appID string, c *config.OAuthConfiguration) *StateCodec {
	return &StateCodec{
		AppID:       appID,
		OAuthConfig: c,
	}
}

func (s *StateCodec) makeStandardClaims() jwt.StandardClaims {
	return jwt.StandardClaims{
		Audience:  s.AppID,
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute).Unix(),
	}
}

func (s *StateCodec) isValidStandardClaims(claims jwt.StandardClaims) bool {
	err := claims.Valid()
	if err != nil {
		return false
	}
	ok := claims.VerifyAudience(s.AppID, true)
	if !ok {
		return false
	}
	return true
}

func (s *StateCodec) EncodeState(state State) (string, error) {
	claims := stateClaims{
		state,
		s.makeStandardClaims(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.OAuthConfig.StateJWTSecret))
}

func (s *StateCodec) DecodeState(encodedState string) (*State, error) {
	claims := stateClaims{}
	_, err := jwt.ParseWithClaims(encodedState, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected JWT alg")
		}
		return []byte(s.OAuthConfig.StateJWTSecret), nil
	})
	if err != nil {
		return nil, NewSSOFailed(InvalidParams, "invalid sso state")
	}
	ok := s.isValidStandardClaims(claims.StandardClaims)
	if !ok {
		return nil, NewSSOFailed(InvalidParams, "invalid sso state")
	}
	return &claims.State, nil
}
