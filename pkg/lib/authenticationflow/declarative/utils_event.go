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
	var assertedAuthenticators []model.Authenticator
	var assertedIdentities []model.Identity

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

	auths, err := collectAuthenticators(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	assertedAuthenticators = make([]model.Authenticator, len(auths))
	for i, info := range auths {
		assertedAuthenticators[i] = info.ToModel()
	}

	ids, err := collectIdentities(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	assertedIdentities = make([]model.Identity, len(ids))
	for i, info := range ids {
		assertedIdentities[i] = info.ToModel()
	}

	return &event.AuthenticationContext{
		AuthenticationFlow:     authenticationFlow,
		User:                   u,
		AMR:                    amr,
		AssertedAuthenticators: assertedAuthenticators,
		AssertedIdentities:     assertedIdentities,
	}, nil
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

func collectIdentities(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (identities []*identity.Info, err error) {
	err = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDoUseIdentity); ok {
				identities = append(identities, n.MilestoneDoUseIdentity())
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if i, ok := intent.(MilestoneDoUseIdentity); ok {
				identities = append(identities, i.MilestoneDoUseIdentity())
			}
			return nil
		},
	}, flows.Root)
	if err != nil {
		return
	}

	// Deduplicate by Info.ID
	seen := map[string]struct{}{}
	var dedupedIdentities []*identity.Info
	for _, info := range identities {
		if _, ok := seen[info.ID]; ok {
			continue
		}
		seen[info.ID] = struct{}{}
		dedupedIdentities = append(dedupedIdentities, info)
	}
	identities = dedupedIdentities

	return
}
