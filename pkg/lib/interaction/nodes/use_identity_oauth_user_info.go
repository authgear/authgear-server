package nodes

import (
	"crypto/subtle"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityOAuthUserInfo{})
}

type InputUseIdentityOAuthUserInfo interface {
	GetProviderAlias() string
	GetCode() string
	GetScope() string
	GetError() string
	GetErrorDescription() string
}

type EdgeUseIdentityOAuthUserInfo struct {
	IsCreating       bool
	Config           config.OAuthSSOProviderConfig
	HashedNonce      string
	ErrorRedirectURI string
}

func (e *EdgeUseIdentityOAuthUserInfo) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentityOAuthUserInfo
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	alias := input.GetProviderAlias()
	nonceSource := ctx.Nonces.GetAndClear()
	code := input.GetCode()
	state := ctx.WebSessionID
	scope := input.GetScope()
	oauthError := input.GetError()
	errorDescription := input.GetErrorDescription()
	hashedNonce := e.HashedNonce

	if e.Config.Alias != alias {
		return nil, fmt.Errorf("interaction: unexpected provider alias %s != %s", e.Config.Alias, alias)
	}

	oauthProvider := ctx.OAuthProviderFactory.NewOAuthProvider(alias)
	if oauthProvider == nil {
		return nil, interaction.ErrOAuthProviderNotFound
	}

	// Handle provider error
	if oauthError != "" {
		return nil, fmt.Errorf("%s: %s", oauthError, errorDescription)
	}

	if nonceSource == "" {
		return nil, fmt.Errorf("nonce does not present in the request")
	}
	nonce := crypto.SHA256String(nonceSource)
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(nonce)) != 1 {
		return nil, fmt.Errorf("invalid nonce")
	}

	userInfo, err := oauthProvider.GetAuthInfo(
		sso.OAuthAuthorizationResponse{
			Code:  code,
			State: state,
			Scope: scope,
		},
		sso.GetAuthInfoParam{
			Nonce: hashedNonce,
		},
	)
	if err != nil {
		return nil, err
	}

	providerID := userInfo.ProviderConfig.ProviderID()
	spec := &identity.Spec{
		Type: authn.IdentityTypeOAuth,
		Claims: map[string]interface{}{
			identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
			identity.IdentityClaimOAuthSubjectID:    userInfo.ProviderUserInfo.ID,
			identity.IdentityClaimOAuthProfile:      userInfo.ProviderRawProfile,
			identity.IdentityClaimOAuthClaims:       userInfo.ProviderUserInfo.ClaimsValue(),
		},
	}

	return &NodeUseIdentityOAuthUserInfo{
		IsCreating:   e.IsCreating,
		IdentitySpec: spec,
	}, nil
}

type NodeUseIdentityOAuthUserInfo struct {
	IsCreating   bool           `json:"is_creating"`
	IdentitySpec *identity.Spec `json:"identity_spec"`
}

func (n *NodeUseIdentityOAuthUserInfo) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityOAuthUserInfo) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUseIdentityOAuthUserInfo) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	if n.IsCreating {
		return []interaction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	}
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
}
