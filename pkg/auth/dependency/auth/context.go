package auth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func IsValidAuthn(ctx context.Context) bool {
	return authn.IsValidAuthn(ctx)
}

func GetUser(ctx context.Context) *authinfo.AuthInfo {
	return authn.GetUser(ctx)
}

func GetSession(ctx context.Context) Session {
	// All session types used in auth conform to our Session interface as well.
	s := authn.GetSession(ctx)
	if s == nil {
		return nil
	}
	return s.(Session)
}
