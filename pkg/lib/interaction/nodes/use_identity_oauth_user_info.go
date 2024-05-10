package nodes

import (
	"crypto/subtle"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityOAuthUserInfo{})
}

type InputUseIdentityOAuthUserInfo interface {
	GetProviderAlias() string
	GetCode() string
	GetError() string
	GetErrorDescription() string
	GetErrorURI() string
}

type EdgeUseIdentityOAuthUserInfo struct {
	IsAuthentication bool
	IsCreating       bool
	Config           oauthrelyingparty.ProviderConfig
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
	oauthError := input.GetError()
	errorDescription := input.GetErrorDescription()
	errorURI := input.GetErrorURI()
	hashedNonce := e.HashedNonce

	providerConfigAlias := e.Config.Alias()
	if providerConfigAlias != alias {
		return nil, fmt.Errorf("interaction: unexpected provider alias %s != %s", providerConfigAlias, alias)
	}

	// Handle provider error
	if oauthError != "" {
		return nil, oauthrelyingpartyutil.NewOAuthError(oauthError, errorDescription, errorURI)
	}

	if nonceSource == "" {
		return nil, fmt.Errorf("nonce does not present in the request")
	}

	nonce := crypto.SHA256String(nonceSource)
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(nonce)) != 1 {
		return nil, fmt.Errorf("invalid nonce")
	}

	oauthProvider := ctx.OAuthProviderFactory.NewOAuthProvider(alias)
	if oauthProvider == nil {
		return nil, api.ErrOAuthProviderNotFound
	}

	redirectURI := ctx.OAuthRedirectURIBuilder.SSOCallbackURL(alias)

	userInfo, err := oauthProvider.GetUserProfile(
		oauthrelyingparty.GetUserProfileOptions{
			Code:        code,
			RedirectURI: redirectURI.String(),
			Nonce:       hashedNonce,
		},
	)
	if err != nil {
		return nil, err
	}

	providerConfig := oauthProvider.Config()
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
