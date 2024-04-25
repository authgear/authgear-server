package declarative

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
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

func resolveAccountLinkingConfigOAuth(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	request *CreateIdentityRequestOAuth) (*config.AuthenticationFlowAccountLinkOAuthItem, error) {
	cfg, err := resolveAccountLinkingConfig(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	var match *config.AuthenticationFlowAccountLinkOAuthItem

	for _, oauthConfig := range cfg.OAuth {
		oauthConfig := oauthConfig
		if oauthConfig.Alias == request.Alias {
			match = oauthConfig
			break
		}
	}

	if match == nil {
		// By default, always error on email conflict
		match = config.DefaultAuthenticationFlowAccountLinkOAuthItem
	}

	return match, nil
}

func linkByOAuthIncomingOAuthSpec(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	request *CreateIdentityRequestOAuth) (cfg config.AccountLinkingConfigObject, conflicts []*identity.Info, err error) {

	oauthConfig, err := resolveAccountLinkingConfigOAuth(ctx, deps, flows, request)
	if err != nil {
		return nil, nil, err
	}

	value, traverseErr := oauthConfig.OAuthClaim.Pointer.Traverse(request.Spec.OAuth.StandardClaims)
	if traverseErr != nil {
		// If we failed to obtain value using the json pointer, just treat it as empty
		value = ""
	}

	valueStr, ok := value.(string)
	if !ok {
		// If value is not string, treat it as empty
		valueStr = ""
	}

	// If value is empty or doesn't exist, no conflicts should occur
	if valueStr == "" {
		return oauthConfig, []*identity.Info{}, nil
	}

	conflicts, err = deps.Identities.ListByClaimJSONPointer(oauthConfig.UserProfile.Pointer, valueStr)
	if err != nil {
		return oauthConfig, nil, err
	}

	// check for identitical identities
	for _, conflict := range conflicts {
		conflict := conflict
		if conflict.Type != model.IdentityTypeOAuth {
			// Not the same type, so must be not identical
			continue
		}
		if !conflict.OAuth.ProviderID.Equal(&request.Spec.OAuth.ProviderID) {
			// Not the same provider
			continue
		}
		if conflict.OAuth.ProviderSubjectID == request.Spec.OAuth.SubjectID {
			// The identity is identical, throw error directly
			spec := request.Spec
			otherSpec := conflict.ToSpec()
			return oauthConfig, nil, identityFillDetails(api.ErrDuplicatedIdentity, spec, &otherSpec)
		}
	}

	return oauthConfig, conflicts, nil
}

type CreateIdentityRequest struct {
	Type model.IdentityType `json:"type,omitempty"`

	LoginID *CreateIdentityRequestLoginID `json:"login_id,omitempty"`
	OAuth   *CreateIdentityRequestOAuth   `json:"oauth,omitempty"`
}

type CreateIdentityRequestOAuth struct {
	Alias string         `json:"alias,omitempty"`
	Spec  *identity.Spec `json:"spec,omitempty"`
}

type CreateIdentityRequestLoginID struct {
	Spec *identity.Spec `json:"spec,omitempty"`
}

func NewCreateOAuthIdentityRequest(alias string, spec *identity.Spec) *CreateIdentityRequest {
	return &CreateIdentityRequest{
		Type: model.IdentityTypeOAuth,
		OAuth: &CreateIdentityRequestOAuth{
			Alias: alias,
			Spec:  spec,
		},
	}
}

func NewCreateLoginIDIdentityRequest(spec *identity.Spec) *CreateIdentityRequest {
	return &CreateIdentityRequest{
		Type: model.IdentityTypeLoginID,
		LoginID: &CreateIdentityRequestLoginID{
			Spec: spec,
		},
	}
}
