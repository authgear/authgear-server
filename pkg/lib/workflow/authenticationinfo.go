package workflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
)

type AuthenticationInfoEntryGetter interface {
	GetAuthenticationInfoEntry(ctx context.Context, deps *Dependencies, flows Workflows) *authenticationinfo.Entry
}
