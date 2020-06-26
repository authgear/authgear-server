package sso

import (
	"encoding/json"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/jwkutil"
	"github.com/skygeario/skygear-server/pkg/jwtutil"
)

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type StateCodec struct {
	AppID       config.AppID
	Clock       clock.Clock
	Credentials *config.JWTKeyMaterials
}

func (s *StateCodec) makeStandardClaims() jwt.Token {
	claims := jwt.New()
	claims.Set(jwt.AudienceKey, string(s.AppID))
	claims.Set(jwt.ExpirationKey, s.Clock.NowUTC().Add(5*time.Minute).Unix())
	return claims
}

func (s *StateCodec) isValidStandardClaims(claims jwt.Token) bool {
	err := jwt.Verify(claims,
		jwt.WithAudience(string(s.AppID)),
		jwt.WithClock(jwtClock{s.Clock}),
	)
	if err != nil {
		return false
	}
	return true
}

func (s *StateCodec) EncodeState(state State) (out string, err error) {
	claims := s.makeStandardClaims()
	claims.Set("state", state)

	key, err := jwkutil.ExtractOctetKey(&s.Credentials.Set, "")
	if err != nil {
		return
	}

	compact, err := jwtutil.Sign(claims, jwa.HS256, key)
	if err != nil {
		return
	}

	out = string(compact)
	return
}

func (s *StateCodec) DecodeState(encodedState string) (*State, error) {
	compact := []byte(encodedState)
	_, payload, err := jwtutil.SplitWithoutVerify(compact)
	if err != nil {
		return nil, NewSSOFailed(InvalidParams, "invalid sso state")
	}

	key, err := jwkutil.ExtractOctetKey(&s.Credentials.Set, "")
	if err != nil {
		return nil, err
	}

	_, err = jws.Verify(compact, jwa.HS256, key)
	if err != nil {
		return nil, NewSSOFailed(InvalidParams, "invalid sso state")
	}

	ok := s.isValidStandardClaims(payload)
	if !ok {
		return nil, NewSSOFailed(InvalidParams, "invalid sso state")
	}

	type stateWrapper struct {
		State State `json:"state"`
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var wrapper stateWrapper
	err = json.Unmarshal(bytes, &wrapper)
	if err != nil {
		return nil, err
	}

	return &wrapper.State, nil
}
