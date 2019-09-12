package authnsession

import (
	"errors"
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
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

var ErrUnknownMFAEnforcement = errors.New("unknown MFA enforcement")
var ErrUnknownClaims = errors.New("unknown claims")
var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	jwt.StandardClaims
	AuthnSession auth.AuthnSession `json:"authn_session"`
}

func NewAuthnSessionToken(secret string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseAuthnSessionToken(secret string, tokenString string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := t.Claims.(*Claims)
	if !ok {
		return nil, ErrUnknownClaims
	}
	if !t.Valid {
		return nil, ErrInvalidToken
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

func NewAuthenticationSessionError(token string, step auth.AuthnSessionStep) skyerr.Error {
	return skyerr.NewErrorWithInfo(
		skyerr.AuthenticationSession,
		"Authentication Session",
		map[string]interface{}{
			"token": token,
			"step":  step,
		},
	)
}

func (p *providerImpl) NewWithToken(token string) (*auth.AuthnSession, error) {
	claims, err := ParseAuthnSessionToken(p.authenticationSessionConfiguration.Secret, token)
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
		return nil, ErrUnknownMFAEnforcement
	}
	return steps, nil
}

func (p *providerImpl) NewFromScratch(userID string, principalID string, reason event.SessionCreateReason) (*auth.AuthnSession, error) {
	clientID := p.authContextGetter.AccessKey().ClientID
	requiredSteps, err := p.getRequiredSteps(userID)
	if err != nil {
		return nil, err
	}
	// Identity is considered finished here.
	finishedSteps := requiredSteps[:1]
	return &auth.AuthnSession{
		ClientID:            clientID,
		UserID:              userID,
		PrincipalID:         principalID,
		RequiredSteps:       requiredSteps,
		FinishedSteps:       finishedSteps,
		SessionCreateReason: string(reason),
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

		sess, err := p.sessionProvider.Create(&authnSess)
		if err != nil {
			return nil, err
		}

		sessionModel := authSession.Format(sess)
		err = p.hookProvider.DispatchEvent(
			event.SessionCreateEvent{
				Reason:   event.SessionCreateReason(authnSess.SessionCreateReason),
				User:     user,
				Identity: identity,
				Session:  sessionModel,
			},
			&user,
		)
		if err != nil {
			return nil, err
		}

		resp := model.NewAuthResponse(user, identity, sess, authnSess.AuthenticatorBearerToken)

		// Refetch the authInfo
		err = p.authInfoStore.GetAuth(authnSess.UserID, &authInfo)
		if err != nil {
			return nil, err
		}

		// Update LastLoginAt and LastSeenAt
		now := p.timeProvider.NowUTC()
		authInfo.LastLoginAt = &now
		authInfo.LastSeenAt = &now
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
	token, err := NewAuthnSessionToken(p.authenticationSessionConfiguration.Secret, claims)
	if err != nil {
		return nil, err
	}
	authnSessionErr := NewAuthenticationSessionError(token, step)
	return authnSessionErr, nil
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
		case skyerr.Error:
			handler.WriteResponse(w, handler.APIResponse{Err: v})
		default:
			panic("unknown response")
		}
	} else {
		handler.WriteResponse(w, handler.APIResponse{Err: skyerr.MakeError(err)})
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

func (p *providerImpl) ResolveUserID(authContext auth.ContextGetter, authnSessionToken string, options ResolveUserIDOptions) (userID string, err error) {
	// Simple case
	authInfo := authContext.AuthInfo()
	if authInfo != nil {
		userID = authInfo.ID
		return
	}

	if authnSessionToken == "" {
		err = skyerr.NewNotAuthenticatedErr()
		return
	}

	authnSession, err := p.NewWithToken(authnSessionToken)
	if err != nil {
		return
	}

	step, ok := authnSession.NextStep()
	if !ok {
		err = skyerr.NewNotAuthenticatedErr()
		return
	}

	switch step {
	case auth.AuthnSessionStepMFA:
		switch options.MFACase {
		case ResolveUserIDMFACaseAlwaysAccept:
			userID = authnSession.UserID
			return
		case ResolveUserIDMfaCaseOnlyWhenNoAuthenticators:
			var authenticators []interface{}
			authenticators, err = p.mfaProvider.ListAuthenticators(userID)
			if err != nil {
				return
			}
			if len(authenticators) > 0 {
				err = skyerr.NewNotAuthenticatedErr()
				return
			}
			userID = authnSession.UserID
			return
		}
	}

	err = skyerr.NewNotAuthenticatedErr()
	return
}
