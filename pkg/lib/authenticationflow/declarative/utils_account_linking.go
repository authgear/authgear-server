package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func resolveAccountLinkingConfig(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (*config.AuthenticationFlowAccountLinking, error) {
	var config *config.AuthenticationFlowAccountLinking

	err := authflow.TraverseIntentFromEndToRoot(func(intent authflow.Intent) error {
		milestone, ok := intent.(MilestoneAccountLinkingConfigGetter)
		if !ok || config != nil {
			return nil
		}
		cfg, err := milestone.MilestoneAccountLinkingConfigGetter(deps)
		if err != nil {
			return err
		}
		config = cfg
		return nil
	}, flows.Root)

	if err != nil {
		return nil, err
	}

	if config == nil {
		return deps.Config.AuthenticationFlow.DefaultAccountLinking, nil
	}

	return config.Merge(deps.Config.AuthenticationFlow.DefaultAccountLinking), nil
}
