package nodes

import (
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
}

type EdgeUseIdentityOAuthProvider struct {
	IsCreating bool
	Configs    []config.OAuthSSOProviderConfig
}

func (e *EdgeUseIdentityOAuthProvider) GetIdentityCandidates() []identity.Candidate {
	candidates := make([]identity.Candidate, len(e.Configs))
	for i, c := range e.Configs {
		conf := c
		candidates[i] = identity.NewOAuthCandidate(&conf)
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
		if c.Alias == alias {
			conf := c
			oauthConfig = &conf
			break
		}
	}
	if oauthConfig == nil {
		panic("interaction: no OAuth provider with specified alias")
	}

	nonceSource := ctx.Nonces.GenerateAndSet()
	errorRedirectURI := input.GetErrorRedirectURI()
	state := ctx.WebSessionID

	oauthProvider := ctx.OAuthProviderFactory.NewOAuthProvider(alias)
	if oauthProvider == nil {
		return nil, interaction.ErrOAuthProviderNotFound
	}

	nonce := crypto.SHA256String(nonceSource)

	param := sso.GetAuthURLParam{
		State: state,
		Nonce: nonce,
	}

	redirectURI, err := oauthProvider.GetAuthURL(param)
	if err != nil {
		return nil, err
	}

	return &NodeUseIdentityOAuthProvider{
		IsCreating:       e.IsCreating,
		Config:           *oauthConfig,
		HashedNonce:      nonce,
		ErrorRedirectURI: errorRedirectURI,
		RedirectURI:      redirectURI,
	}, nil
}

type NodeUseIdentityOAuthProvider struct {
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
			IsCreating:       n.IsCreating,
			Config:           n.Config,
			HashedNonce:      n.HashedNonce,
			ErrorRedirectURI: n.ErrorRedirectURI,
		},
	}, nil
}
