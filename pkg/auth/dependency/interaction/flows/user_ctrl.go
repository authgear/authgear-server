package flows

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	oauthprotocol "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

type TokenIssuer interface {
	IssueTokens(
		client config.OAuthClientConfiguration,
		attrs *authn.Attrs,
	) (auth.AuthSession, oauthprotocol.TokenResponse, error)
}

type UserProvider interface {
	Get(id string) (*model.User, error)
}

type UserController struct {
	AuthInfos           authinfo.Store
	Users               UserProvider
	TokenIssuer         TokenIssuer
	SessionCookieConfig session.CookieConfiguration
	Sessions            session.Provider
	Hooks               hook.Provider
	Time                time.Provider
	Clients             []config.OAuthClientConfiguration
}

func (c *UserController) makeResponse(attrs *authn.Attrs) (*model.AuthResponse, error) {
	user, err := c.Users.Get(attrs.UserID)
	if err != nil {
		return nil, err
	}

	resp := &model.AuthResponse{}
	resp.User = *user
	return resp, nil
}

func (c *UserController) CreateSession(
	i *interaction.Interaction,
	ir *interaction.Result,
) (*AuthResult, error) {
	resp, err := c.makeResponse(ir.Attrs)
	if err != nil {
		return nil, err
	}
	result := &AuthResult{Response: resp}

	session, token := c.Sessions.MakeSession(ir.Attrs)
	err = c.Sessions.Create(session)
	if err != nil {
		return nil, err
	}

	result.Cookies = []*http.Cookie{c.SessionCookieConfig.NewCookie(token)}

	result.Response.SessionID = session.SessionID()

	identity := ir.Identity.ToModel()
	reason := auth.SessionCreateReasonLogin
	if intent, ok := i.Intent.(*interaction.IntentLogin); ok {
		if intent.OriginalIntentType == interaction.IntentTypeSignup {
			reason = auth.SessionCreateReasonSignup
		}
	}

	err = c.Hooks.DispatchEvent(
		event.SessionCreateEvent{
			Reason:   string(reason),
			User:     result.Response.User,
			Identity: identity,
			Session:  *session.ToAPIModel(),
		},
		&result.Response.User,
	)
	if err != nil {
		return nil, err
	}

	err = c.updateLoginTime(ir.Attrs.UserID)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *UserController) MakeAuthResult(attrs *authn.Attrs) (*AuthResult, error) {
	resp, err := c.makeResponse(attrs)
	if err != nil {
		return nil, err
	}
	result := &AuthResult{Response: resp}
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
