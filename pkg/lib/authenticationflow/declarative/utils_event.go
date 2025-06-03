package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func GetAuthenticationContext(ctx context.Context, flows authenticationflow.Flows, deps *authenticationflow.Dependencies) (*event.AuthenticationContext, error) {
	authCtx := &event.AuthenticationContext{}

	flowRef := authenticationflow.FindCurrentFlowReference(flows.Root)
	if flowRef != nil {
		authCtx.AuthenticationFlow = &event.AuthenticationFlowContext{
			Type: string(flowRef.Type),
			Name: flowRef.Name,
		}
	}

	userID, err := getUserID(flows)
	if err == nil {
		u, err := deps.Users.Get(ctx, userID, accesscontrol.RoleGreatest)
		if err != nil && !errors.Is(err, user.ErrUserNotFound) {
			return nil, err
		}
		authCtx.User = u
	}

	amr, err := collectAMR(ctx, deps, flows)
	if err != nil {
		return nil, err
	}
	authCtx.AMR = amr

	auths, err := collectAuthenticators(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	assertedAuthenticators := make([]model.Authenticator, len(auths))
	for i, info := range auths {
		assertedAuthenticators[i] = info.ToModel()
	}
	authCtx.AssertedAuthenticators = assertedAuthenticators

	return authCtx, nil
}

func collectAuthenticators(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (authenticators []*authenticator.Info, err error) {
	err = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDidAuthenticate); ok {
				if info, ok := n.MilestoneDidAuthenticateAuthenticator(); ok {
					authenticators = append(authenticators, info)
				}
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if i, ok := intent.(MilestoneDidAuthenticate); ok {
				if info, ok := i.MilestoneDidAuthenticateAuthenticator(); ok {
					authenticators = append(authenticators, info)
				}
			}
			return nil
		},
	}, flows.Root)

	if err != nil {
		return
	}

	// Deduplicate by Info.ID
	seen := map[string]struct{}{}
	var dedupedAuthenticators []*authenticator.Info
	for _, info := range authenticators {
		if _, ok := seen[info.ID]; ok {
			continue
		}
		seen[info.ID] = struct{}{}
		dedupedAuthenticators = append(dedupedAuthenticators, info)
	}
	authenticators = dedupedAuthenticators

	return
}
