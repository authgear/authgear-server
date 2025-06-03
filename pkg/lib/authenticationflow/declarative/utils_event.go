package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func GetAuthenticationContext(ctx context.Context, flows authenticationflow.Flows, deps *authenticationflow.Dependencies) (*event.AuthenticationContext, error) {
	var authenticationFlow *event.AuthenticationFlowContext
	var u *model.User
	var amr []string

	flowRef := authenticationflow.FindCurrentFlowReference(flows.Root)
	if flowRef != nil {
		authenticationFlow = &event.AuthenticationFlowContext{
			Type: string(flowRef.Type),
			Name: flowRef.Name,
		}
	}

	userID, err := getUserID(flows)
	if err == nil {
		u, err = deps.Users.Get(ctx, userID, accesscontrol.RoleGreatest)
		if err != nil && !errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
	}

	amr, err = collectAMR(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	assertedAuthenticators, err := collectAssertedAuthenticators(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	assertedIdentities, err := collectAssertedIdentities(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	authCtx := &event.AuthenticationContext{
		AuthenticationFlow:     authenticationFlow,
		User:                   u,
		AMR:                    amr,
		AssertedAuthenticators: []model.Authenticator{},
		AssertedIdentities:     []model.Identity{},
	}

	for _, authn := range assertedAuthenticators {
		authCtx.AddAssertedAuthenticator(authn.ToModel())
	}
	for _, iden := range assertedIdentities {
		authCtx.AddAssertedIdentity(iden.ToModel())
	}

	return authCtx, nil
}

func collectAssertedAuthenticators(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticators []*authenticator.Info, err error) {
	err = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDidAuthenticate); ok {
				if info, ok := n.MilestoneDidAuthenticateAuthenticator(); ok {
					authenticators = append(authenticators, info)
				}
			}
			if n, ok := nodeSimple.(MilestoneDoCreateAuthenticator); ok {
				authenticators = append(authenticators, n.MilestoneDoCreateAuthenticator())
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if i, ok := intent.(MilestoneDidAuthenticate); ok {
				if info, ok := i.MilestoneDidAuthenticateAuthenticator(); ok {
					authenticators = append(authenticators, info)
				}
			}
			if i, ok := intent.(MilestoneDoCreateAuthenticator); ok {
				authenticators = append(authenticators, i.MilestoneDoCreateAuthenticator())
			}
			return nil
		},
	}, flows.Root)

	if err != nil {
		return
	}

	return
}

func collectAssertedIdentities(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (identities []*identity.Info, err error) {
	err = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDoUseIdentity); ok {
				identities = append(identities, n.MilestoneDoUseIdentity())
			}
			if n, ok := nodeSimple.(MilestoneDoCreateIdentity); ok {
				identities = append(identities, n.MilestoneDoCreateIdentity())
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if i, ok := intent.(MilestoneDoUseIdentity); ok {
				identities = append(identities, i.MilestoneDoUseIdentity())
			}
			if i, ok := intent.(MilestoneDoCreateIdentity); ok {
				identities = append(identities, i.MilestoneDoCreateIdentity())
			}
			return nil
		},
	}, flows.Root)
	if err != nil {
		return
	}

	return
}
