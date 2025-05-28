package authenticationflow

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
)

type IdentitySpecGetter interface {
	GetIdentitySpecs(ctx context.Context, deps *Dependencies, flows Flows) []*identity.Spec
}

func CollectIdentitySpecs(ctx context.Context, deps *Dependencies, flows Flows) (identitySpecs []*identity.Spec, err error) {
	err = TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
			if n, ok := nodeSimple.(IdentitySpecGetter); ok {
				c := n.GetIdentitySpecs(ctx, deps, flows.Replace(w))
				identitySpecs = append(identitySpecs, c...)
			}

			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			if i, ok := intent.(IdentitySpecGetter); ok {
				c := i.GetIdentitySpecs(ctx, deps, flows.Replace(w))
				identitySpecs = append(identitySpecs, c...)
			}

			return nil
		},
	}, flows.Nearest)
	if err != nil {
		return
	}

	return
}
