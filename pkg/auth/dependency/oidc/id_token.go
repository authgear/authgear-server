package oidc

import (
	gotime "time"

	"github.com/dgrijalva/jwt-go"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/urlprefix"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type UserClaims struct {
	jwt.StandardClaims
	User      *model.User `json:"skygear_user,omitempty"`
	SessionID string      `json:"skygear_session_id,omitempty"`
}

type IDTokenClaims struct {
	UserClaims
	Nonce string `json:"nonce,omitempty"`
}

type IDTokenIssuer struct {
	OIDCConfig       config.OIDCConfiguration
	URLPrefix        urlprefix.Provider
	AuthInfoStore    authinfo.Store
	UserProfileStore userprofile.Store
	Time             time.Provider
}

// IDTokenValidDuration is the valid period of ID token.
// It can be short, since id_token_hint should accept expired ID tokens.
const IDTokenValidDuration = 5 * gotime.Minute

func (ti *IDTokenIssuer) IssueIDToken(client config.OAuthClientConfiguration, session auth.AuthSession, nonce string) (string, error) {
	userClaims, err := ti.LoadUserClaims(session)
	if err != nil {
		return "", err
	}

	now := ti.Time.NowUTC()
	userClaims.StandardClaims.Audience = client.ClientID()
	userClaims.StandardClaims.IssuedAt = now.Unix()
	userClaims.StandardClaims.ExpiresAt = now.Add(IDTokenValidDuration).Unix()

	claims := &IDTokenClaims{
		UserClaims: *userClaims,
		Nonce:      nonce,
	}

	key := ti.OIDCConfig.Keys[0]
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(key.PrivateKey))
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = key.KID
	return token.SignedString(privKey)
}

func (ti *IDTokenIssuer) LoadUserClaims(session auth.AuthSession) (*UserClaims, error) {
	allowProfile := false
	for _, scope := range oauth.SessionScopes(session) {
		if scope == oauth.FullAccessScope {
			allowProfile = true
		}
	}

	claims := &UserClaims{
		StandardClaims: jwt.StandardClaims{
			Issuer:  ti.URLPrefix.Value().String(),
			Subject: session.AuthnAttrs().UserID,
		},
	}

	if !allowProfile {
		return claims, nil
	}

	authInfo := &authinfo.AuthInfo{}
	if err := ti.AuthInfoStore.GetAuth(session.AuthnAttrs().UserID, authInfo); err != nil {
		return nil, err
	}

	userProfile, err := ti.UserProfileStore.GetUserProfile(session.AuthnAttrs().UserID)
	if err != nil {
		return nil, err
	}

	user := model.NewUser(*authInfo, userProfile)
	claims.User = &user
	claims.SessionID = session.SessionID()

	return claims, nil
}
