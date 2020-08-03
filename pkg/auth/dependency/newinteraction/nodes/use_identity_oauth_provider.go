package nodes

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/crypto"
)

func init() {
	newinteraction.RegisterNode(&NodeUseIdentityOAuthProvider{})
}

type InputUseIdentityOAuthProvider interface {
	GetProviderAlias() string
	GetState() string
	GetNonceSource() *http.Cookie
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

func (e *EdgeUseIdentityOAuthProvider) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputUseIdentityOAuthProvider)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
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

	nonceSource := input.GetNonceSource()
	errorRedirectURI := input.GetErrorRedirectURI()
	state := input.GetState()

	oauthProvider := ctx.OAuthProviderFactory.NewOAuthProvider(alias)
	if oauthProvider == nil {
		return nil, newinteraction.ErrOAuthProviderNotFound
	}

	nonce := crypto.SHA256String(nonceSource.Value)

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

func (n *NodeUseIdentityOAuthProvider) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseIdentityOAuthProvider) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeUseIdentityOAuthUserInfo{
			IsCreating:       n.IsCreating,
			Config:           n.Config,
			HashedNonce:      n.HashedNonce,
			ErrorRedirectURI: n.ErrorRedirectURI,
		},
	}, nil
}
