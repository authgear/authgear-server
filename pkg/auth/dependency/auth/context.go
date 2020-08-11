package auth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/core/authn"
)

func IsValidAuthn(ctx context.Context) bool {
	return authn.IsValidAuthn(ctx)
}

func GetUserID(ctx context.Context) *string {
	return authn.GetUserID(ctx)
}

func GetSession(ctx context.Context) AuthSession {
	// All session types used in auth conform to our Session interface as well.
	s := authn.GetSession(ctx)
	if s == nil {
		return nil
	}
	return s.(AuthSession)
}
