package authn

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/session"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

func GetSession(ctx context.Context) SessionContainer {
	sctx := session.GetContext(ctx)
	if sctx == nil || sctx.Session == nil {
		return nil
	}
	return sctx.Session
}

func GetUser(ctx context.Context) *authinfo.AuthInfo {
	sctx := session.GetContext(ctx)
	if sctx == nil || sctx.User == nil {
		return nil
	}
	return sctx.User
}
