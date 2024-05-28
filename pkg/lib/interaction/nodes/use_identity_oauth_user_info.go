package nodes

import (
	"crypto/subtle"
	"fmt"

	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityOAuthUserInfo{})
}

type InputUseIdentityOAuthUserInfo interface {
	GetProviderAlias() string
	GetQuery() string
}

type EdgeUseIdentityOAuthUserInfo struct {
	IsAuthentication bool
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
	query := input.GetQuery()
	nonceSource := ctx.Nonces.GetAndClear()
	hashedNonce := e.HashedNonce

	providerConfigAlias := e.Config.Alias()
	if providerConfigAlias != alias {
		return nil, fmt.Errorf("interaction: unexpected provider alias %s != %s", providerConfigAlias, alias)
	}

	if nonceSource == "" {
		return nil, fmt.Errorf("nonce does not present in the request")
	}

	nonce := crypto.SHA256String(nonceSource)
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(nonce)) != 1 {
		return nil, fmt.Errorf("invalid nonce")
	}

	redirectURI := ctx.OAuthRedirectURIBuilder.SSOCallbackURL(alias)

	providerConfig, err := ctx.OAuthProviderFactory.GetProviderConfig(alias)
	if err != nil {
		return nil, err
	}

	userInfo, err := ctx.OAuthProviderFactory.GetUserProfile(
		alias,
		oauthrelyingparty.GetUserProfileOptions{
			Query:       query,
			RedirectURI: redirectURI.String(),
			Nonce:       hashedNonce,
		},
	)
	if err != nil {
		return nil, err
	}

	providerID := providerConfig.ProviderID()
	spec := &identity.Spec{
		Type: model.IdentityTypeOAuth,
		OAuth: &identity.OAuthSpec{
			ProviderID:     providerID,
			SubjectID:      userInfo.ProviderUserID,
			RawProfile:     userInfo.ProviderRawProfile,
			StandardClaims: userInfo.StandardAttributes,
		},
	}

	return &NodeUseIdentityOAuthUserInfo{
		IsAuthentication: e.IsAuthentication,
		IsCreating:       e.IsCreating,
		IdentitySpec:     spec,
	}, nil
}

type NodeUseIdentityOAuthUserInfo struct {
	IsAuthentication bool           `json:"is_authentication"`
	IsCreating       bool           `json:"is_creating"`
	IdentitySpec     *identity.Spec `json:"identity_spec"`
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
	return []interaction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec, IsAuthentication: n.IsAuthentication}}, nil
}
