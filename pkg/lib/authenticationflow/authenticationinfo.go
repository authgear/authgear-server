package authenticationflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
)

type AuthenticationInfoEntryGetter interface {
	GetAuthenticationInfoEntry(ctx context.Context, deps *Dependencies, flows Flows) *authenticationinfo.Entry
}

func GetAuthenticationInfoEntry(ctx context.Context, deps *Dependencies, flows Flows) (*authenticationinfo.Entry, bool) {
	var e *authenticationinfo.Entry
	_ = TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
			if n, ok := nodeSimple.(AuthenticationInfoEntryGetter); ok {
				e = n.GetAuthenticationInfoEntry(ctx, deps, flows)
			}

			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			if i, ok := intent.(AuthenticationInfoEntryGetter); ok {
				e = i.GetAuthenticationInfoEntry(ctx, deps, flows)
			}

			return nil
		},
	}, flows.Nearest)
	if e != nil {
		return e, true
	}
	return nil, false
}
