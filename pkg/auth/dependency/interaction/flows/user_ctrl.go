package flows

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/auth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/interaction"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type UserProvider interface {
	Get(id string) (*model.User, error)
	UpdateLoginTime(user *model.User, lastLoginAt time.Time) error
}

type HookProvider interface {
	DispatchEvent(payload event.Payload, user *model.User) error
}

type SessionProvider interface {
	MakeSession(*authn.Attrs) (*session.IDPSession, string)
	Create(*session.IDPSession) error
}

type UserController struct {
	Users         UserProvider
	SessionCookie session.CookieDef
	Sessions      SessionProvider
	Hooks         HookProvider
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

	err = c.Users.UpdateLoginTime(&result.Response.User, session.CreatedAt)
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

	return result, nil
}
