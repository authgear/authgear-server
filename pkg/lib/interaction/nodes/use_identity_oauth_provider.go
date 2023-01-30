package nodes

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

func init() {
	interaction.RegisterNode(&NodeUseIdentityOAuthProvider{})
}

type InputUseIdentityOAuthProvider interface {
	GetProviderAlias() string
	GetErrorRedirectURI() string
	GetPrompt() []string
}

type EdgeUseIdentityOAuthProvider struct {
	IsAuthentication bool
	IsCreating       bool
	Configs          []config.OAuthSSOProviderConfig
	FeatureConfig    *config.OAuthSSOProvidersFeatureConfig
}

func (e *EdgeUseIdentityOAuthProvider) GetIdentityCandidates() []identity.Candidate {
	candidates := []identity.Candidate{}
	for _, c := range e.Configs {
		conf := c
		if !identity.IsOAuthSSOProviderTypeDisabled(conf.Type, e.FeatureConfig) {
			candidates = append(candidates, identity.NewOAuthCandidate(&conf))
		}
	}
	return candidates
}

func (e *EdgeUseIdentityOAuthProvider) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputUseIdentityOAuthProvider
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	alias := input.GetProviderAlias()
	var oauthConfig *config.OAuthSSOProviderConfig
	for _, c := range e.Configs {
		if identity.IsOAuthSSOProviderTypeDisabled(c.Type, e.FeatureConfig) {
			continue
		}
		if c.Alias == alias {
			conf := c
			oauthConfig = &conf
			break
		}
	}
	if oauthConfig == nil {
		return nil, api.ErrOAuthProviderNotFound
	}

	nonceSource := ctx.Nonces.GenerateAndSet()
	errorRedirectURI := input.GetErrorRedirectURI()
	state := ctx.WebSessionID

	oauthProvider := ctx.OAuthProviderFactory.NewOAuthProvider(alias)
	if oauthProvider == nil {
		return nil, api.ErrOAuthProviderNotFound
	}

	nonce := crypto.SHA256String(nonceSource)

	param := sso.GetAuthURLParam{
		State:  state,
		Nonce:  nonce,
		Prompt: input.GetPrompt(),
	}

	redirectURI, err := oauthProvider.GetAuthURL(param)
	if err != nil {
		return nil, err
	}

	return &NodeUseIdentityOAuthProvider{
		IsAuthentication: e.IsAuthentication,
		IsCreating:       e.IsCreating,
		Config:           *oauthConfig,
		HashedNonce:      nonce,
		ErrorRedirectURI: errorRedirectURI,
		RedirectURI:      redirectURI,
	}, nil
}

type NodeUseIdentityOAuthProvider struct {
	IsAuthentication bool                          `json:"is_authentication"`
	IsCreating       bool                          `json:"is_creating"`
	Config           config.OAuthSSOProviderConfig `json:"provider_config"`
	HashedNonce      string                        `json:"hashed_nonce"`
	ErrorRedirectURI string                        `json:"error_redirect_uri"`
	RedirectURI      string                        `json:"redirect_uri"`
}

// GetRedirectURI implements RedirectURIGetter.
func (n *NodeUseIdentityOAuthProvider) GetRedirectURI() string {
	return n.RedirectURI
}

// GetErrorRedirectURI implements ErrorRedirectURIGetter.
func (n *NodeUseIdentityOAuthProvider) GetErrorRedirectURI() string {
	return n.ErrorRedirectURI
}

func (n *NodeUseIdentityOAuthProvider) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeUseIdentityOAuthProvider) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeUseIdentityOAuthProvider) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeUseIdentityOAuthUserInfo{
			IsAuthentication: n.IsAuthentication,
			IsCreating:       n.IsCreating,
			Config:           n.Config,
			HashedNonce:      n.HashedNonce,
			ErrorRedirectURI: n.ErrorRedirectURI,
		},
	}, nil
}
