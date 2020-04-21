package flows

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	oauthprotocol "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coremodel "github.com/skygeario/skygear-server/pkg/core/model"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type AuthAPITokenIssuer interface {
	IssueAuthAPITokens(
		client config.OAuthClientConfiguration,
		attrs *authn.Attrs,
	) (auth.AuthSession, oauthprotocol.TokenResponse, error)
}

type UserController struct {
	AuthInfos           authinfo.Store
	UserProfiles        userprofile.Store
	TokenIssuer         AuthAPITokenIssuer
	SessionCookieConfig session.CookieConfiguration
	Sessions            session.Provider
	Hooks               hook.Provider
	Time                time.Provider
	Clients             []config.OAuthClientConfiguration
}

func (c *UserController) makeResponse(attrs *authn.Attrs) (*model.AuthResponse, error) {
	resp := &model.AuthResponse{}
	var authInfo authinfo.AuthInfo
	err := c.AuthInfos.GetAuth(attrs.UserID, &authInfo)
	if err != nil {
		return nil, err
	}

	userProfile, err := c.UserProfiles.GetUserProfile(attrs.UserID)
	if err != nil {
		return nil, err
	}

	resp.User = model.NewUser(authInfo, userProfile)
	identity := model.NewIdentityFromAttrs(attrs)
	resp.Identity = &identity

	return resp, nil
}

func (c *UserController) CreateSession(
	i *interaction.Interaction,
	attrs *authn.Attrs,
	isAuthAPI bool,
) (*AuthResult, error) {
	client, ok := coremodel.GetClientConfig(c.Clients, i.ClientID)
	if !ok && isAuthAPI {
		return nil, interaction.ErrInvalidCredentials
	}

	resp, err := c.makeResponse(attrs)
	if err != nil {
		return nil, err
	}
	result := &AuthResult{Response: resp}

	var authSession auth.AuthSession
	if isAuthAPI && !client.AuthAPIUseCookie() {
		session, resp, err := c.TokenIssuer.IssueAuthAPITokens(
			client, attrs,
		)
		if err != nil {
			return nil, err
		}

		authSession = session
		result.Response.AccessToken = resp.GetAccessToken()
		result.Response.RefreshToken = resp.GetRefreshToken()
		result.Response.ExpiresIn = resp.GetExpiresIn()
	} else {
		session, token := c.Sessions.MakeSession(attrs)
		err = c.Sessions.Create(session)
		if err != nil {
			return nil, err
		}

		authSession = session
		result.Cookies = []*http.Cookie{c.SessionCookieConfig.NewCookie(token)}
	}

	result.Response.SessionID = authSession.SessionID()

	reason := auth.SessionCreateReasonLogin
	if i.Intent.Type() == interaction.IntentTypeSignup {
		reason = auth.SessionCreateReasonSignup
	}
	err = c.Hooks.DispatchEvent(
		event.SessionCreateEvent{
			Reason:   string(reason),
			User:     result.Response.User,
			Identity: *result.Response.Identity,
			Session:  *authSession.ToAPIModel(),
		},
		&result.Response.User,
	)
	if err != nil {
		return nil, err
	}

	err = c.updateLoginTime(attrs.UserID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *UserController) updateLoginTime(userID string) error {
	authInfo := &authinfo.AuthInfo{}
	err := c.AuthInfos.GetAuth(userID, authInfo)
	if err != nil {
		return err
	}

	// Update LastLoginAt and LastSeenAt
	now := c.Time.NowUTC()
	authInfo.LastLoginAt = &now
	authInfo.LastSeenAt = &now
	authInfo.RefreshDisabledStatus(now)
	err = c.AuthInfos.UpdateAuth(authInfo)
	if err != nil {
		return err
	}

	return nil
}
