package declarative

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
)

func GetAuthenticationContext(ctx context.Context, deps *authenticationflow.Dependencies, flows authenticationflow.Flows) (*event.AuthenticationContext, error) {
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

	amr, err = CollectAMR(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	assertedAuthentications, err := collectAssertedAuthentications(flows)
	if err != nil {
		return nil, err
	}

	assertedIdentifications, err := collectAssertedIdentifications(flows)
	if err != nil {
		return nil, err
	}

	authCtx := &event.AuthenticationContext{
		AuthenticationFlow:      authenticationFlow,
		User:                    u,
		AMR:                     amr,
		AssertedAuthentications: []model.Authentication{},
		AssertedIdentifications: []model.Identification{},
	}

	for _, authn := range assertedAuthentications {
		authCtx.AddAssertedAuthentication(authn)
	}
	for _, iden := range assertedIdentifications {
		authCtx.AddAssertedIdentification(iden)
	}

	return authCtx, nil
}

func collectAssertedIdentifications(flows authenticationflow.Flows) (identifications []model.Identification, err error) {
	err = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDoUseIdentity); ok {
				identifications = append(identifications, n.MilestoneDoUseIdentityIdentification())
			}
			if n, ok := nodeSimple.(MilestoneDoCreateIdentity); ok {
				identifications = append(identifications, n.MilestoneDoCreateIdentityIdentification())
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if i, ok := intent.(MilestoneDoUseIdentity); ok {
				identifications = append(identifications, i.MilestoneDoUseIdentityIdentification())
			}
			if i, ok := intent.(MilestoneDoCreateIdentity); ok {
				identifications = append(identifications, i.MilestoneDoCreateIdentityIdentification())
			}
			return nil
		},
	}, flows.Root)
	if err != nil {
		return
	}

	return
}

func collectAssertedAuthentications(flows authenticationflow.Flows) (authens []model.Authentication, err error) {
	err = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDidAuthenticate); ok {
				if a, ok := n.MilestoneDidAuthenticateAuthentication(); ok {
					authens = append(authens, *a)
				}
			}
			if n, ok := nodeSimple.(MilestoneDoCreateAuthenticator); ok {
				authn, ok := n.MilestoneDoCreateAuthenticatorAuthentication()
				if ok {
					authens = append(authens, *authn)
				}
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if i, ok := intent.(MilestoneDidAuthenticate); ok {
				if a, ok := i.MilestoneDidAuthenticateAuthentication(); ok {
					authens = append(authens, *a)
				}
			}
			if i, ok := intent.(MilestoneDoCreateAuthenticator); ok {
				authn, ok := i.MilestoneDoCreateAuthenticatorAuthentication()
				if ok {
					authens = append(authens, *authn)
				}
			}
			return nil
		},
	}, flows.Root)

	if err != nil {
		return
	}

	return
}

func IsPreAuthenticatedTriggered(flows authenticationflow.Flows) (triggered bool) {
	_ = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if _, ok := nodeSimple.(MilestonePreAuthenticated); ok {
				triggered = true
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if _, ok := intent.(MilestonePreAuthenticated); ok {
				triggered = true
			}
			return nil
		},
	}, flows.Root)
	return
}

func toRateLimitWeights(eventratelimits event.RateLimits) ratelimit.Weights {
	if eventratelimits == nil {
		return nil
	}

	weights := ratelimit.Weights{}
	for rl, rlRequirements := range eventratelimits {
		weights[ratelimit.RateLimitGroup(rl)] = rlRequirements.Weight
	}
	return weights
}
