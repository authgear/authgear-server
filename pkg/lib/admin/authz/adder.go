package authz

import (
	"fmt"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

type Adder struct {
	Clock clock.Clock
}

func (a *Adder) AddAuthz(
	auth config.AdminAPIAuth,
	appID config.AppID,
	authKey *config.AdminAPIAuthKey,
	auditContext interface{},
	hdr http.Header) (err error) {
	switch auth {
	case config.AdminAPIAuthNone:
		break
	case config.AdminAPIAuthJWT:
		if authKey == nil {
			panic("authz: no key configured for admin API auth")
		}

		now := a.Clock.NowUTC()
		payload := jwt.New()
		_ = payload.Set(jwt.AudienceKey, string(appID))
		_ = payload.Set(jwt.IssuedAtKey, now.Unix())
		_ = payload.Set(jwt.ExpirationKey, now.Add(duration.Short).Unix())
		if auditContext != nil {
			_ = payload.Set(JWTKeyAuditContext, auditContext)
		}

		key, _ := authKey.Set.Key(0)

		var token []byte
		token, err = jwtutil.Sign(payload, jwa.RS256, key)
		if err != nil {
			return
		}

		hdr.Set("Authorization", fmt.Sprintf("Bearer %s", string(token)))
	}

	return
}
