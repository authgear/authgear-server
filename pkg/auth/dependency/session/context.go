package session

import (
	"context"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

type Context struct {
	Session *Session
	User    *authinfo.AuthInfo
}

func (ctx *Context) ToAuthnInfo(now time.Time) *authn.Info {
	if ctx == nil {
		return nil
	}
	if ctx.Session == nil || ctx.User == nil {
		return &authn.Info{IsValid: false}
	}

	return &authn.Info{
		IsValid:                        true,
		UserID:                         ctx.User.ID,
		UserVerified:                   ctx.User.IsVerified(),
		UserDisabled:                   ctx.User.IsDisabled(now),
		SessionIdentityID:              ctx.Session.PrincipalID,
		SessionIdentityType:            ctx.Session.PrincipalType,
		SessionIdentityUpdatedAt:       ctx.Session.PrincipalUpdatedAt,
		SessionAuthenticatorID:         ctx.Session.AuthenticatorID,
		SessionAuthenticatorType:       ctx.Session.AuthenticatorType,
		SessionAuthenticatorOOBChannel: ctx.Session.AuthenticatorOOBChannel,
		SessionAuthenticatorUpdatedAt:  ctx.Session.AuthenticatorUpdatedAt,
	}
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

func WithSession(ctx context.Context, s *Session, u *authinfo.AuthInfo) context.Context {
	sCtx := &Context{s, u}
	return context.WithValue(ctx, contextKey, sCtx)
}

func GetContext(ctx context.Context) *Context {
	sCtx, _ := ctx.Value(contextKey).(*Context)
	return sCtx
}
