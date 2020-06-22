package flows

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	oauthprotocol "github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/clock"
	"github.com/skygeario/skygear-server/pkg/core/authn"
	"github.com/skygeario/skygear-server/pkg/core/config"
)

type TokenIssuer interface {
	IssueTokens(
		client config.OAuthClientConfiguration,
		attrs *authn.Attrs,
	) (auth.AuthSession, oauthprotocol.TokenResponse, error)
}

type UserProvider interface {
	Get(id string) (*model.User, error)
	UpdateLoginTime(user *model.User, lastLoginAt time.Time) error
}

type HookProvider interface {
	DispatchEvent(payload event.Payload, user *model.User) error
}

type UserController struct {
	Users         UserProvider
	TokenIssuer   TokenIssuer
	SessionCookie session.CookieDef
	Sessions      session.Provider
	Hooks         HookProvider
	Clock         clock.Clock
	Clients       []config.OAuthClientConfiguration
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

	result.Cookies = []*http.Cookie{c.SessionCookie.New(token)}

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

	now := c.Clock.NowUTC()
	err = c.Users.UpdateLoginTime(&result.Response.User, now)
	if err != nil {
		return nil, err
	}

	return result, nil
}
