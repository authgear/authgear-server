package session

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/patrickmn/go-cache"

	"github.com/authgear/authgear-server/pkg/api/model"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

var simpleCache = cache.New(5*time.Minute, 10*time.Minute)

const cacheKeyJWKs = "jwks"

type jwtClock struct {
	Clock clock.Clock
}

func (c jwtClock) Now() time.Time {
	return c.Clock.NowUTC()
}

type SessionInfoMiddleware struct {
	AuthgearConfig *portalconfig.AuthgearConfig
	HTTPClient     HTTPClient
	Clock          clock.Clock
}

func (m *SessionInfoMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch m.AuthgearConfig.WebSDKSessionType {
		case "refresh_token":
			m.handleAuthorizationHeader(next, w, r)
		case "cookie":
			fallthrough
		default:
			m.handleCookie(next, w, r)
		}
	})
}

func (m *SessionInfoMiddleware) handleAuthorizationHeader(next http.Handler, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	jwkSet, err := m.getJWKs(ctx)
	if err != nil {
		panic(err)
	}

	var sessionInfo *model.SessionInfo
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		// keep sessionInfo as nil. It means no session.
	} else {
		sessionInfo = m.jwtToSessionInfo(jwkSet, r.Header)
	}

	r = r.WithContext(WithSessionInfo(ctx, sessionInfo))
	next.ServeHTTP(w, r)
}

func (m *SessionInfoMiddleware) jwtToSessionInfo(jwkSet jwk.Set, header http.Header) (sessionInfo *model.SessionInfo) {
	// Initialize to zero value.
	// Zero value means invalid session.
	sessionInfo = &model.SessionInfo{}

	token, err := jwt.ParseHeader(header, "Authorization",
		jwt.WithVerify(true),
		jwt.WithKeySet(jwkSet),
		jwt.WithClock(jwtClock{m.Clock}),
		jwt.WithAcceptableSkew(duration.ClockSkew),
	)
	if err != nil {
		return
	}

	sessionInfo.UserID = token.Subject()

	anonymousIface, ok := token.Get(string(model.ClaimUserIsAnonymous))
	if !ok {
		panic(fmt.Errorf("expected claim to be present: %v", model.ClaimUserIsAnonymous))
	}
	sessionInfo.UserAnonymous = anonymousIface.(bool)

	isVerifiedIface, ok := token.Get(string(model.ClaimUserIsVerified))
	if !ok {
		panic(fmt.Errorf("expected claim to be present: %v", model.ClaimUserIsVerified))
	}
	sessionInfo.UserVerified = isVerifiedIface.(bool)

	canReauthenticate, ok := token.Get(string(model.ClaimUserCanReauthenticate))
	if !ok {
		panic(fmt.Errorf("expected claim to be present: %v", model.ClaimUserCanReauthenticate))
	}
	sessionInfo.UserCanReauthenticate = canReauthenticate.(bool)

	// auth_time is newly added to at+jwt, so it may not be present.
	if authTimeIface, ok := token.Get(string(model.ClaimAuthTime)); ok {
		switch v := authTimeIface.(type) {
		case float64:
			sessionInfo.AuthenticatedAt = time.Unix(int64(v), 0).UTC()
		case int64:
			sessionInfo.AuthenticatedAt = time.Unix(v, 0).UTC()
		default:
			panic(fmt.Errorf("unexpected type: %v %T", model.ClaimAuthTime, authTimeIface))
		}
	}

	// amr is newly added to at+jwt, so it may not be present.
	if amrIface, ok := token.Get(string(model.ClaimAMR)); ok {
		amrSlice := amrIface.([]interface{})
		for _, amrValue := range amrSlice {
			amrStr := amrValue.(string)
			sessionInfo.SessionAMR = append(sessionInfo.SessionAMR, amrStr)
		}
	}

	rolesIface, ok := token.Get(string(model.ClaimAuthgearRoles))
	if !ok {
		panic(fmt.Errorf("expected claim to be present: %v", model.ClaimAuthgearRoles))
	}
	rolesSlice := rolesIface.([]interface{})
	for _, roleIface := range rolesSlice {
		role := roleIface.(string)
		sessionInfo.EffectiveRoles = append(sessionInfo.EffectiveRoles, role)
	}

	sessionInfo.IsValid = true
	return
}

func (m *SessionInfoMiddleware) getJWKs(ctx context.Context) (jwk.Set, error) {
	jwkIface, ok := simpleCache.Get(cacheKeyJWKs)
	if ok {
		return jwkIface.(jwk.Set), nil
	}

	parsedEndpoint, err := url.Parse(m.AuthgearConfig.Endpoint)
	if err != nil {
		return nil, err
	}

	// HTTP Host header includes port, so we use .Host
	httpHostHeader := parsedEndpoint.Host

	connectionEndpoint := parsedEndpoint.String()
	if m.AuthgearConfig.EndpointInternal != "" {
		connectionEndpoint = m.AuthgearConfig.EndpointInternal
	}

	jwksURI, err := url.JoinPath(connectionEndpoint, "/oauth2/jwks")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", jwksURI, nil)
	if err != nil {
		return nil, err
	}
	// This is the most important line.
	req.Host = httpHostHeader

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch OIDC JWKs: unexpected status code: %v", resp.StatusCode)
	}

	jwkSet, err := jwk.ParseReader(resp.Body)
	if err != nil {
		return nil, err
	}
	simpleCache.Set(cacheKeyJWKs, jwkSet, 0)

	return jwkSet, nil
}

func (m *SessionInfoMiddleware) handleCookie(next http.Handler, w http.ResponseWriter, r *http.Request) {
	sessionInfo, err := model.NewSessionInfoFromHeaders(r.Header)
	if err != nil {
		panic(err)
	}

	r = r.WithContext(WithSessionInfo(r.Context(), sessionInfo))
	next.ServeHTTP(w, r)
}
