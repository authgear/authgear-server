package authnsession

import (
	"net/http"
	gotime "time"

	"github.com/dgrijalva/jwt-go"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	authSession "github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/errors"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/time"
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

type Claims struct {
	jwt.StandardClaims
	AuthnSession auth.AuthnSession `json:"authn_session"`
}

func newAuthnSessionToken(secret string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func parseAuthnSessionToken(secret string, tokenString string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected JWT alg")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, errInvalidToken
	}
	claims, ok := t.Claims.(*Claims)
	if !ok {
		return nil, errInvalidToken
	}
	if !t.Valid {
		return nil, errInvalidToken
	}
	return claims, nil
}

type providerImpl struct {
	authContextGetter                  auth.ContextGetter
	mfaConfiguration                   config.MFAConfiguration
	authenticationSessionConfiguration config.AuthenticationSessionConfiguration
	timeProvider                       time.Provider
	mfaProvider                        mfa.Provider
	authInfoStore                      authinfo.Store
	sessionProvider                    session.Provider
	sessionWriter                      session.Writer
	identityProvider                   principal.IdentityProvider
	hookProvider                       hook.Provider
	userProfileStore                   userprofile.Store
}

func NewProvider(
	authContextGetter auth.ContextGetter,
	mfaConfiguration config.MFAConfiguration,
	authenticationSessionConfiguration config.AuthenticationSessionConfiguration,
	timeProvider time.Provider,
	mfaProvider mfa.Provider,
	authInfoStore authinfo.Store,
	sessionProvider session.Provider,
	sessionWriter session.Writer,
	identityProvider principal.IdentityProvider,
	hookProvider hook.Provider,
	userProfileStore userprofile.Store,
) Provider {
	return &providerImpl{
		authContextGetter:                  authContextGetter,
		mfaConfiguration:                   mfaConfiguration,
		authenticationSessionConfiguration: authenticationSessionConfiguration,
		timeProvider:                       timeProvider,
		mfaProvider:                        mfaProvider,
		authInfoStore:                      authInfoStore,
		sessionProvider:                    sessionProvider,
		sessionWriter:                      sessionWriter,
		identityProvider:                   identityProvider,
		hookProvider:                       hookProvider,
		userProfileStore:                   userProfileStore,
	}
}

func NewAuthenticationSessionError(token string, step auth.AuthnSessionStep) error {
	return AuthenticationSessionRequired.NewWithDetails(
		"authentication session is required",
		skyerr.Details{
			"token": skyerr.APIErrorString(token),
			"step":  skyerr.APIErrorString(step),
		},
	)
}

func (p *providerImpl) NewFromToken(token string) (*auth.AuthnSession, error) {
	claims, err := parseAuthnSessionToken(p.authenticationSessionConfiguration.Secret, token)
	if err != nil {
		return nil, err
	}
	return &claims.AuthnSession, nil
}

func (p *providerImpl) getRequiredSteps(userID string) ([]auth.AuthnSessionStep, error) {
	steps := []auth.AuthnSessionStep{auth.AuthnSessionStepIdentity}
	enforcement := p.mfaConfiguration.Enforcement
	switch enforcement {
	case config.MFAEnforcementOptional:
		authenticators, err := p.mfaProvider.ListAuthenticators(userID)
		if err != nil {
			return nil, err
		}
		if len(authenticators) > 0 {
			steps = append(steps, auth.AuthnSessionStepMFA)
		}
	case config.MFAEnforcementRequired:
		steps = append(steps, auth.AuthnSessionStepMFA)
	case config.MFAEnforcementOff:
		break
	default:
		return nil, errors.New("unknown MFA enforcement")
	}
	return steps, nil
}

func (p *providerImpl) NewFromScratch(userID string, prin principal.Principal, reason auth.SessionCreateReason) (*auth.AuthnSession, error) {
	now := p.timeProvider.NowUTC()
	clientID := p.authContextGetter.AccessKey().ClientID
	requiredSteps, err := p.getRequiredSteps(userID)
	if err != nil {
		return nil, errors.HandledWithMessage(err, "cannot get required authn steps")
	}
	// Identity is considered finished here.
	finishedSteps := requiredSteps[:1]
	return &auth.AuthnSession{
		ClientID:            clientID,
		UserID:              userID,
		PrincipalID:         prin.PrincipalID(),
		PrincipalType:       auth.PrincipalType(prin.ProviderID()),
		PrincipalUpdatedAt:  now,
		RequiredSteps:       requiredSteps,
		FinishedSteps:       finishedSteps,
		SessionCreateReason: reason,
	}, nil
}

func (p *providerImpl) GenerateResponseAndUpdateLastLoginAt(authnSess auth.AuthnSession) (interface{}, error) {
	step, ok := authnSess.NextStep()
	if !ok {
		var authInfo authinfo.AuthInfo
		err := p.authInfoStore.GetAuth(authnSess.UserID, &authInfo)
		if err != nil {
			return nil, err
		}

		userProfile, err := p.userProfileStore.GetUserProfile(authnSess.UserID)
		if err != nil {
			return nil, err
		}

		user := model.NewUser(authInfo, userProfile)

		prin, err := p.identityProvider.GetPrincipalByID(authnSess.PrincipalID)
		if err != nil {
			return nil, err
		}
		identity := model.NewIdentity(p.identityProvider, prin)

		sess, tokens, err := p.sessionProvider.Create(&authnSess)
		if err != nil {
			return nil, err
		}

		sessionModel := authSession.Format(sess)
		err = p.hookProvider.DispatchEvent(
			event.SessionCreateEvent{
				Reason:   auth.SessionCreateReason(authnSess.SessionCreateReason),
				User:     user,
				Identity: identity,
				Session:  sessionModel,
			},
			&user,
		)
		if err != nil {
			return nil, err
		}

		resp := model.NewAuthResponse(user, identity, tokens, authnSess.AuthenticatorBearerToken)

		// Refetch the authInfo
		err = p.authInfoStore.GetAuth(authnSess.UserID, &authInfo)
		if err != nil {
			return nil, err
		}

		// Update LastLoginAt and LastSeenAt
		now := p.timeProvider.NowUTC()
		authInfo.LastLoginAt = &now
		authInfo.LastSeenAt = &now
		authInfo.RefreshDisabledStatus()
		err = p.authInfoStore.UpdateAuth(&authInfo)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
	now := p.timeProvider.NowUTC()
	expiresAt := now.Add(5 * gotime.Minute)
	claims := Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			IssuedAt:  now.Unix(),
		},
		AuthnSession: authnSess,
	}
	token, err := newAuthnSessionToken(p.authenticationSessionConfiguration.Secret, claims)
	if err != nil {
		return nil, err
	}
	authnSessionErr := NewAuthenticationSessionError(token, step)
	return authnSessionErr, nil
}

func (p *providerImpl) GenerateResponseWithSession(sess *auth.Session, mfaBearerToken string) (interface{}, error) {
	var authInfo authinfo.AuthInfo
	err := p.authInfoStore.GetAuth(sess.UserID, &authInfo)
	if err != nil {
		return nil, err
	}

	userProfile, err := p.userProfileStore.GetUserProfile(sess.UserID)
	if err != nil {
		return nil, err
	}
	user := model.NewUser(authInfo, userProfile)

	prin, err := p.identityProvider.GetPrincipalByID(sess.PrincipalID)
	if err != nil {
		return nil, err
	}
	identity := model.NewIdentity(p.identityProvider, prin)

	resp := model.NewAuthResponse(user, identity, auth.SessionTokens{ID: sess.ID}, mfaBearerToken)
	return resp, nil
}

func (p *providerImpl) WriteResponse(w http.ResponseWriter, resp interface{}, err error) {
	if err == nil {
		switch v := resp.(type) {
		case model.AuthResponse:
			// Do not touch the cookie if it is not in the response.
			if v.MFABearerToken == "" {
				p.sessionWriter.WriteSession(w, &v.AccessToken, nil)
			} else {
				p.sessionWriter.WriteSession(w, &v.AccessToken, &v.MFABearerToken)
			}
			handler.WriteResponse(w, handler.APIResponse{Result: v})
		case error:
			handler.WriteResponse(w, handler.APIResponse{Error: skyerr.AsAPIError(v)})
		default:
			panic("authnsession: unknown response")
		}
	} else {
		if skyerr.IsKind(err, mfa.InvalidBearerToken) {
			p.sessionWriter.ClearMFABearerToken(w)
		}
		handler.WriteResponse(w, handler.APIResponse{Error: skyerr.AsAPIError(err)})
	}
}

func (p *providerImpl) AlterResponse(w http.ResponseWriter, resp interface{}, err error) interface{} {
	if v, ok := resp.(model.AuthResponse); ok && err == nil {
		// Do not touch the cookie if it is not in the response.
		if v.MFABearerToken == "" {
			p.sessionWriter.WriteSession(w, &v.AccessToken, nil)
		} else {
			p.sessionWriter.WriteSession(w, &v.AccessToken, &v.MFABearerToken)
		}
		return v
	}
	return resp
}

func (p *providerImpl) Resolve(authContext auth.ContextGetter, authnSessionToken string, options ResolveOptions) (userID string, sess *auth.Session, authnSession *auth.AuthnSession, err error) {
	// Simple case
	sess, _ = authContext.Session()
	if sess != nil {
		userID = sess.UserID
		return
	}

	if authnSessionToken == "" {
		err = authz.NewNotAuthenticatedError()
		return
	}

	authnSession, err = p.NewFromToken(authnSessionToken)
	if err != nil {
		return
	}

	step, ok := authnSession.NextStep()
	if !ok {
		err = errInvalidToken
		return
	}

	switch step {
	case auth.AuthnSessionStepMFA:
		switch options.MFAOption {
		case ResolveMFAOptionAlwaysAccept:
			userID = authnSession.UserID
			return
		case ResolveMFAOptionOnlyWhenNoAuthenticators:
			var authenticators []mfa.Authenticator
			authenticators, err = p.mfaProvider.ListAuthenticators(authnSession.UserID)
			if err != nil {
				return
			}
			if len(authenticators) > 0 {
				err = authz.NewNotAuthenticatedError()
				return
			}
			userID = authnSession.UserID
			return
		}
	}

	err = errors.New("unexpected authn session state")
	return
}
