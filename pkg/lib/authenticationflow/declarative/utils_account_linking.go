package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
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

func resolveAccountLinkingConfigOAuth(cfg *config.AuthenticationFlowAccountLinking, request *CreateIdentityRequestOAuth) *config.AccountLinkingOAuth {
	var match *config.AccountLinkingOAuth

	for _, oauthConfig := range cfg.OAuth {
		oauthConfig := oauthConfig
		if oauthConfig.Alias == request.Alias {
			match = oauthConfig
			break
		}
	}

	if match == nil {
		// By default, always error on email conflict
		match = &config.AccountLinkingOAuth{
			OAuthClaim:  config.AccountLinkingJSONPointer{Pointer: jsonpointer.MustParse("/email")},
			UserProfile: config.AccountLinkingJSONPointer{Pointer: jsonpointer.MustParse("/email")},
			Action:      config.AccountLinkingOAuthActionError,
		}
	}

	return match
}

func linkByOAuthIncomingOAuthSpec(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	request *CreateIdentityRequestOAuth) (conflicts []*identity.Info, err error) {

	accountLinkingConfig, err := resolveAccountLinkingConfig(ctx, deps, flows)
	config := resolveAccountLinkingConfigOAuth(accountLinkingConfig, request)

	value, traverseErr := config.OAuthClaim.Pointer.Traverse(request.Spec.OAuth.StandardClaims)
	if traverseErr != nil {
		// If we failed to obtain value using the json pointer, just treat it as empty
		value = ""
	}

	valueStr, ok := value.(string)
	if !ok {
		// If value is not string, treat it as empty
		valueStr = ""
	}

	conflicts, err = deps.Identities.ListByClaimJSONPointer(config.UserProfile.Pointer, valueStr)
	if err != nil {
		return nil, err
	}

	// TODO(tung): This function should exclude identical identities

	return conflicts, nil
}
