package session

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type Context interface {
	Session() *Session
	User() *authinfo.AuthInfo
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

type sessionContext struct {
	s *Session
	u *authinfo.AuthInfo
}

func (c *sessionContext) Session() *Session {
	return c.s
}

func (c *sessionContext) User() *authinfo.AuthInfo {
	return c.u
}

func WithSession(ctx context.Context, s *Session, u *authinfo.AuthInfo) context.Context {
	sCtx := &sessionContext{s, u}
	return context.WithValue(ctx, contextKey, sCtx)
}

func GetContext(ctx context.Context) Context {
	return ctx.Value(contextKey).(*sessionContext)
}
