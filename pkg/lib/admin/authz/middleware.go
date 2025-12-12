package authz

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

var authzHeader = regexp.MustCompile("^Bearer (.*)$")

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

var MiddlewareLogger = slogutil.NewLogger("admin-api-authz")

type Middleware struct {
	Auth    config.AdminAPIAuth
	AppID   config.AppID
	AuthKey *config.AdminAPIAuthKey
	Clock   clock.Clock
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := MiddlewareLogger.GetLogger(ctx)
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
				logger.Debug(ctx, "invalid authz header", slog.String("header", r.Header.Get("Authorization")))
				break
			}
			token, err := jwt.ParseString(match[1], jwt.WithKeySet(keySet), jwt.WithValidate(false))
			if err != nil {
				logger.WithError(err).Debug(ctx, "invalid token")
				break
			}

			err = jwt.Validate(token,
				jwt.WithClock(&jwtClock{m.Clock}),
				jwt.WithAudience(string(m.AppID)),
			)
			if err != nil {
				logger.WithError(err).Debug(ctx, "invalid token")
				break
			}

			authorized = true

			auditCtx, ok := token.Get(JWTKeyAuditContext)
			if ok {
				if auditCtx, ok := auditCtx.(map[string]any); ok {
					ctx = WithAdminAuthzAudit(ctx, auditCtx)
					r = r.WithContext(ctx)
				} else {
					logger.WithError(err).Error(ctx, "invalid audit_context, ignoring")
				}
			}
		}

		if !authorized {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
