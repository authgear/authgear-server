package oauth

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
)

type AuthorizationHandler struct {
	Context context.Context
}

func (h *AuthorizationHandler) Handle(r protocol.AuthorizationRequest) AuthorizationResult {
	// TODO(oauth): handle authz request
	return nil
}
