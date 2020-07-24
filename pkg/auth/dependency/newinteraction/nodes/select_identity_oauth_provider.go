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
	newinteraction.RegisterNode(&NodeSelectIdentityOAuthProvider{})
}

type InputSelectIdentityOAuthProvider interface {
	GetProviderAlias() string
	GetState() string
	GetNonceSource() *http.Cookie
	GetErrorRedirectURI() string
}

type EdgeSelectIdentityOAuthProvider struct {
	Config config.OAuthSSOProviderConfig
}

func (e *EdgeSelectIdentityOAuthProvider) GetIdentityCandidate() identity.Candidate {
	return identity.NewOAuthCandidate(&e.Config)
}

func (e *EdgeSelectIdentityOAuthProvider) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityOAuthProvider)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	alias := input.GetProviderAlias()
	if e.Config.Alias != alias {
		return nil, newinteraction.ErrIncompatibleInput
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

	return &NodeSelectIdentityOAuthProvider{
		Config:           e.Config,
		HashedNonce:      nonce,
		ErrorRedirectURI: errorRedirectURI,
		RedirectURI:      redirectURI,
	}, nil
}

type NodeSelectIdentityOAuthProvider struct {
	Config           config.OAuthSSOProviderConfig `json:"provider_config"`
	HashedNonce      string                        `json:"hashed_nonce"`
	ErrorRedirectURI string                        `json:"error_redirect_uri"`
	RedirectURI      string                        `json:"redirect_uri"`
}

// GetErrorRedirectURI implements ErrorRedirectURIGetter
func (n *NodeSelectIdentityOAuthProvider) GetErrorRedirectURI() string {
	return n.ErrorRedirectURI
}

func (n *NodeSelectIdentityOAuthProvider) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityOAuthProvider) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeSelectIdentityOAuthUserInfo{
			Config:           n.Config,
			HashedNonce:      n.HashedNonce,
			ErrorRedirectURI: n.ErrorRedirectURI,
		},
	}, nil
}
