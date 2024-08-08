package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func resolveAccountLinkingConfigsOAuth(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	request *CreateIdentityRequestOAuth,
) ([]*config.AccountLinkingOAuthItem, error) {
	cfg := deps.Config.AccountLinking

	matches := []*config.AccountLinkingOAuthItem{}

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

func resolveAccountLinkingConfigsLoginID(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	request *CreateIdentityRequestLoginID,
) ([]*config.AccountLinkingLoginIDItem, error) {
	cfg := deps.Config.AccountLinking

	matches := []*config.AccountLinkingLoginIDItem{}

	for _, loginIDConfig := range cfg.LoginID {
		loginIDConfig := loginIDConfig
		if loginIDConfig.Key == request.Spec.LoginID.Key {
			matches = append(matches, loginIDConfig)
		}
	}

	if len(matches) == 0 {
		// Otherwise, use the default based on the login ID type.
		switch request.Spec.LoginID.Type {
		case model.LoginIDKeyTypeEmail:
			matches = append(matches, config.DefaultAccountLinkingLoginIDEmailItem)
		case model.LoginIDKeyTypePhone:
			matches = append(matches, config.DefaultAccountLinkingLoginIDPhoneItem)
		case model.LoginIDKeyTypeUsername:
			matches = append(matches, config.DefaultAccountLinkingLoginIDUsernameItem)
		default:
			panic(fmt.Errorf("unexpected login ID type: %v", request.Spec.LoginID.Type))
		}
	}

	return matches, nil
}

type AccountLinkingConflict struct {
	Identity  *identity.Info              `json:"identity"`
	Action    config.AccountLinkingAction `json:"action"`
	LoginFlow string                      `json:"login_flow"`
}

func newAccountLinkingConflictWithIncomingOAuth(identity *identity.Info, cfg *config.AccountLinkingOAuthItem, overrides *config.AuthenticationFlowAccountLinking) *AccountLinkingConflict {
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

func newAccountLinkingConflictWithIncomingLoginID(iden *identity.Info, cfg *config.AccountLinkingLoginIDItem, overrides *config.AuthenticationFlowAccountLinking) *AccountLinkingConflict {
	conflict := &AccountLinkingConflict{
		Identity:  iden,
		Action:    cfg.Action,
		LoginFlow: "",
	}

	if overrides != nil && cfg.Name != "" {
		var overrideItem *config.AuthenticationFlowAccountLinkingLoginIDItem
		for _, item := range overrides.LoginID {
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

func linkByIncomingOAuthSpec(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	userID string,
	request *CreateIdentityRequestOAuth,
	identificationJSONPointer jsonpointer.T,
) (conflicts []*AccountLinkingConflict, err error) {
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
		value, traverseErr := oauthConfig.OAuthClaim.MustGetOneLevelJSONPointerOrPanic().Traverse(request.Spec.OAuth.StandardClaims)
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

		idenConflicts, err := deps.Identities.ListByClaim(oauthConfig.UserProfile.MustGetFirstLevelReferenceTokenOrPanic(), valueStr)
		if err != nil {
			return nil, err
		}

		for _, iden := range idenConflicts {
			iden := iden

			// Exclude identities that actually belong to this user.
			if iden.UserID == userID {
				continue
			}

			// Exclude duplicates
			if _, exist := conflictedIdentityIDs[iden.ID]; exist {
				continue
			}
			conflictedIdentityIDs[iden.ID] = iden.ID
			conflict := newAccountLinkingConflictWithIncomingOAuth(iden, oauthConfig, configOverride)
			conflicts = append(conflicts, conflict)
		}
	}

	// check for identical identities
	for _, conflict := range conflicts {
		conflict := conflict
		if conflict.Identity.Type != model.IdentityTypeOAuth {
			// Not the same type, so must be not identical
			continue
		}
		if !conflict.Identity.OAuth.ProviderID.Equal(request.Spec.OAuth.ProviderID) {
			// Not the same provider
			continue
		}
		if conflict.Identity.OAuth.ProviderSubjectID == request.Spec.OAuth.SubjectID {
			// The identity is identical, throw error directly
			spec := request.Spec
			otherSpec := conflict.Identity.ToSpec()
			return nil, identity.NewErrDuplicatedIdentity(spec, &otherSpec)
		}
	}

	return conflicts, nil
}

func linkByIncomingLoginIDSpec(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	userID string,
	request *CreateIdentityRequestLoginID,
	identificationJSONPointer jsonpointer.T,
) (conflicts []*AccountLinkingConflict, err error) {
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

	loginIDConfigs, err := resolveAccountLinkingConfigsLoginID(ctx, deps, flows, request)
	if err != nil {
		return nil, err
	}

	normalizedValue, _, err := deps.LoginIDs.CheckAndNormalize(*request.Spec.LoginID)
	if err != nil {
		return nil, err
	}

	// For deduplication
	conflictedIdentityIDs := map[string]interface{}{}

	for _, loginIDConfig := range loginIDConfigs {

		idenConflicts, err := deps.Identities.ListByClaim(
			loginIDConfig.UserProfile.MustGetFirstLevelReferenceTokenOrPanic(),
			normalizedValue,
		)
		if err != nil {
			return nil, err
		}

		for _, iden := range idenConflicts {
			iden := iden

			// Exclude identities that actually belong to this user.
			if iden.UserID == userID {
				continue
			}

			// Exclude duplicates
			if _, exist := conflictedIdentityIDs[iden.ID]; exist {
				continue
			}
			conflictedIdentityIDs[iden.ID] = iden.ID
			conflict := newAccountLinkingConflictWithIncomingLoginID(iden, loginIDConfig, configOverride)
			conflicts = append(conflicts, conflict)
		}
	}

	// check for identical identities
	for _, conflict := range conflicts {
		conflict := conflict
		if conflict.Identity.Type != model.IdentityTypeLoginID {
			// Not the same type, so must be not identical
			continue
		}
		if conflict.Identity.LoginID.LoginIDType != request.Spec.LoginID.Type {
			// Not of the same login ID type.
			continue
		}
		if conflict.Identity.LoginID.LoginIDKey != request.Spec.LoginID.Key {
			// Not of the same login ID key.
			continue
		}
		if conflict.Identity.LoginID.LoginID == normalizedValue {
			// The identity is identical, throw error directly.
			spec := request.Spec
			otherSpec := conflict.Identity.ToSpec()
			return nil, identity.NewErrDuplicatedIdentity(spec, &otherSpec)
		}
	}

	return conflicts, nil
}

type CreateIdentityRequest struct {
	Type model.IdentityType `json:"type,omitempty"`

	LoginID *CreateIdentityRequestLoginID `json:"login_id,omitempty"`
	OAuth   *CreateIdentityRequestOAuth   `json:"oauth,omitempty"`
	LDAP    *CreateIdentityRequestLDAP    `json:"ldap,omitempty"`
}

type CreateIdentityRequestOAuth struct {
	Alias string         `json:"alias,omitempty"`
	Spec  *identity.Spec `json:"spec,omitempty"`
}

type CreateIdentityRequestLoginID struct {
	Spec *identity.Spec `json:"spec,omitempty"`
}

type CreateIdentityRequestLDAP struct {
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

func NewCreateLDAPIdentityRequest(spec *identity.Spec) *CreateIdentityRequest {
	return &CreateIdentityRequest{
		Type: model.IdentityTypeLDAP,
		LDAP: &CreateIdentityRequestLDAP{
			Spec: spec,
		},
	}
}
