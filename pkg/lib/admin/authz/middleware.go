package authz

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
)

var authzHeader = regexp.MustCompile("^Bearer (.*)$")

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("admin-api-authz")}
}

type Middleware struct {
	Logger  Logger
	Auth    config.AdminAPIAuth
	AppID   config.AppID
	AuthKey *config.AdminAPIAuthKey
	Clock   clock.Clock
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized := false
		switch m.Auth {
		case config.AdminAPIAuthNone:
			authorized = true

		case config.AdminAPIAuthJWT:
			if m.AuthKey == nil {
				panic("authz: no key configured for admin API auth")
			}
			keySet, err := jwk.PublicSetOf(m.AuthKey.Set)
			if err != nil {
				panic(fmt.Errorf("authz: cannot extract public keys: %w", err))
			}

			match := authzHeader.FindStringSubmatch(r.Header.Get("Authorization"))
			if len(match) != 2 {
				m.Logger.
					WithField("header", r.Header.Get("Authorization")).
					Debug("invalid authz header")
				break
			}
			token, err := jwt.ParseString(match[1], jwt.WithKeySet(keySet), jwt.WithValidate(false))
			if err != nil {
				m.Logger.
					WithError(err).
					Debug("invalid token")
				break
			}

			err = jwt.Validate(token,
				jwt.WithClock(&jwtClock{m.Clock}),
				jwt.WithAudience(string(m.AppID)),
			)
			if err != nil {
				m.Logger.
					WithError(err).
					Debug("invalid token")
				break
			}

			authorized = true

			requestCtx := r.Context()
			auditCtx, ok := token.Get(JWTKeyAuditContext)
			if ok {
				newCtx := WithAdminAuthzAudit(requestCtx, auditCtx)
				r = r.WithContext(newCtx)
			}
		}

		if !authorized {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
