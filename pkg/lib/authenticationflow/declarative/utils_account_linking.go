package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func resolveAccountLinkingConfigsOAuth(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	request *CreateIdentityRequestOAuth) ([]*config.AccountLinkingOAuthItem, error) {
	cfg := deps.Config.AccountLinking

	var matches []*config.AccountLinkingOAuthItem = []*config.AccountLinkingOAuthItem{}

	for _, oauthConfig := range cfg.OAuth {
		oauthConfig := oauthConfig
		if oauthConfig.Alias == request.Alias {
			matches = append(matches, oauthConfig)
		}
	}

	if len(matches) == 0 {
		// By default, always error on email conflict
		matches = append(matches, config.DefaultAccountLinkingOAuthItem)
	}

	return matches, nil
}

type AccountLinkingConflict struct {
	Identity  *identity.Info              `json:"identity"`
	Action    config.AccountLinkingAction `json:"action"`
	LoginFlow string                      `json:"login_flow"`
}

func newAccountLinkingOAuthConflict(identity *identity.Info, cfg *config.AccountLinkingOAuthItem, overrides *config.AuthenticationFlowAccountLinking) *AccountLinkingConflict {
	conflict := &AccountLinkingConflict{
		Identity:  identity,
		Action:    cfg.Action,
		LoginFlow: "",
	}

	if overrides != nil && cfg.Name != "" {
		var overrideItem *config.AuthenticationFlowAccountLinkingOAuthItem
		for _, item := range overrides.OAuth {
			item := item
			if item.Name != cfg.Name {
				continue
			}
			overrideItem = item
		}
		if overrideItem != nil {
			if overrideItem.Action != "" {
				conflict.Action = overrideItem.Action
			}
			if overrideItem.LoginFlow != "" {
				conflict.LoginFlow = overrideItem.LoginFlow
			}
		}
	}

	return conflict
}

func linkByOAuthIncomingOAuthSpec(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	request *CreateIdentityRequestOAuth,
	identificationJSONPointer jsonpointer.T) (conflicts []*AccountLinkingConflict, err error) {

	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	current, err := authflow.FlowObject(flowRootObject, identificationJSONPointer)
	if err != nil {
		return nil, err
	}

	var configOverride *config.AuthenticationFlowAccountLinking
	configOverrider, ok := current.(config.AuthenticationFlowObjectAccountLinkingConfigProvider)
	if ok {
		configOverride = configOverrider.GetAccountLinkingConfig()
	}

	oauthConfigs, err := resolveAccountLinkingConfigsOAuth(ctx, deps, flows, request)
	if err != nil {
		return nil, err
	}

	// For deduplication
	conflictedIdentityIDs := map[string]interface{}{}

	for _, oauthConfig := range oauthConfigs {
		value, traverseErr := oauthConfig.OAuthClaim.GetJSONPointer().Traverse(request.Spec.OAuth.StandardClaims)
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
			continue
		}

		idenConflicts, err := deps.Identities.ListByClaimJSONPointer(oauthConfig.UserProfile.GetJSONPointer(), valueStr)
		if err != nil {
			return nil, err
		}

		for _, iden := range idenConflicts {
			iden := iden
			// Exclude duplicates
			if _, exist := conflictedIdentityIDs[iden.ID]; exist {
				continue
			}
			conflictedIdentityIDs[iden.ID] = iden.ID
			conflict := newAccountLinkingOAuthConflict(iden, oauthConfig, configOverride)
			conflicts = append(conflicts, conflict)
		}
	}

	// check for identitical identities
	for _, conflict := range conflicts {
		conflict := conflict
		if conflict.Identity.Type != model.IdentityTypeOAuth {
			// Not the same type, so must be not identical
			continue
		}
		if !conflict.Identity.OAuth.ProviderID.Equal(&request.Spec.OAuth.ProviderID) {
			// Not the same provider
			continue
		}
		if conflict.Identity.OAuth.ProviderSubjectID == request.Spec.OAuth.SubjectID {
			// The identity is identical, throw error directly
			spec := request.Spec
			otherSpec := conflict.Identity.ToSpec()
			return nil, identityFillDetails(api.ErrDuplicatedIdentity, spec, &otherSpec)
		}
	}

	return conflicts, nil
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
