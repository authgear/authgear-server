package sso

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/jwkutil"
)

type stateClaims struct {
	State
	jwt.StandardClaims
}

type StateCodec struct {
	AppID       config.AppID
	Credentials *config.JWTKeyMaterials
}

func (s *StateCodec) makeStandardClaims() jwt.StandardClaims {
	return jwt.StandardClaims{
		Audience:  string(s.AppID),
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute).Unix(),
	}
}

func (s *StateCodec) isValidStandardClaims(claims jwt.StandardClaims) bool {
	err := claims.Valid()
	if err != nil {
		return false
	}
	ok := claims.VerifyAudience(string(s.AppID), true)
	if !ok {
		return false
	}
	return true
}

func (s *StateCodec) EncodeState(state State) (out string, err error) {
	claims := stateClaims{
		state,
		s.makeStandardClaims(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	key, err := jwkutil.ExtractOctetKey(&s.Credentials.Set, "")
	if err != nil {
		return
	}

	out, err = token.SignedString(key)
	if err != nil {
		return
	}

	return
}

func (s *StateCodec) DecodeState(encodedState string) (*State, error) {
	claims := stateClaims{}
	_, err := jwt.ParseWithClaims(encodedState, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected JWT alg")
		}

		key, err := jwkutil.ExtractOctetKey(&s.Credentials.Set, "")
		if err != nil {
			return nil, err
		}

		return key, nil
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
