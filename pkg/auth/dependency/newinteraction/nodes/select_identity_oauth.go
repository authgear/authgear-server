package nodes

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/core/crypto"
	"github.com/authgear/authgear-server/pkg/core/skyerr"
)

func init() {
	newinteraction.RegisterNode(&NodeSelectIdentityOAuthProvider{})
	newinteraction.RegisterNode(&NodeSelectIdentityOAuthUserInfo{})
}

var ErrOAuthProviderNotFound = skyerr.NotFound.WithReason("OAuthProviderNotFound").New("oauth provider not found")

type InputSelectIdentityOAuthProvider interface {
	GetProviderAlias() string
	GetState() string
	GetNonceSource() *http.Cookie
	GetErrorRedirectURI() string
}

type EdgeSelectIdentityOAuthProvider struct {
	Config config.OAuthSSOProviderConfig
}

type NodeSelectIdentityOAuthProvider struct {
	Config           config.OAuthSSOProviderConfig `json:"provider_config"`
	HashedNonce      string                        `json:"hashed_nonce"`
	ErrorRedirectURI string                        `json:"error_redirect_uri"`
	RedirectURI      string                        `json:"redirect_uri"`
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
		return nil, ErrOAuthProviderNotFound
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

type InputSelectIdentityOAuthUserInfo interface {
	GetProviderAlias() string
	GetNonceSource() *http.Cookie
	GetCode() string
	GetState() string
	GetScope() string
	GetError() string
	GetErrorDescription() string
}

type EdgeSelectIdentityOAuthUserInfo struct {
	Config           config.OAuthSSOProviderConfig
	HashedNonce      string
	ErrorRedirectURI string
}

func (e *EdgeSelectIdentityOAuthUserInfo) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityOAuthUserInfo)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	alias := input.GetProviderAlias()
	nonceSource := input.GetNonceSource()
	code := input.GetCode()
	state := input.GetState()
	scope := input.GetScope()
	oauthError := input.GetError()
	errorDescription := input.GetErrorDescription()
	hashedNonce := e.HashedNonce

	if e.Config.Alias != alias {
		return nil, fmt.Errorf("interaction: unexpected provider alias %s != %s", e.Config.Alias, alias)
	}

	oauthProvider := ctx.OAuthProviderFactory.NewOAuthProvider(alias)
	if oauthProvider == nil {
		return nil, ErrOAuthProviderNotFound
	}

	// Handle provider error
	if oauthError != "" {
		msg := "login failed"
		if errorDescription != "" {
			msg += ": " + errorDescription
		}
		return nil, sso.NewSSOFailed(sso.SSOUnauthorized, msg)
	}

	if nonceSource == nil || nonceSource.Value == "" {
		return nil, sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
	}
	nonce := crypto.SHA256String(nonceSource.Value)
	if subtle.ConstantTimeCompare([]byte(hashedNonce), []byte(nonce)) != 1 {
		return nil, sso.NewSSOFailed(sso.SSOUnauthorized, "invalid nonce")
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

	return &NodeSelectIdentityOAuthUserInfo{
		UserInfo: userInfo,
	}, nil
}

type NodeSelectIdentityOAuthUserInfo struct {
	UserInfo sso.AuthInfo
}

func (n *NodeSelectIdentityOAuthUserInfo) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityOAuthUserInfo) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	providerID := n.UserInfo.ProviderConfig.ProviderID()
	spec := &identity.Spec{
		Type: authn.IdentityTypeOAuth,
		Claims: map[string]interface{}{
			identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
			identity.IdentityClaimOAuthSubjectID:    n.UserInfo.ProviderUserInfo.ID,
			identity.IdentityClaimOAuthProfile:      n.UserInfo.ProviderRawProfile,
			identity.IdentityClaimOAuthClaims:       n.UserInfo.ProviderUserInfo.ClaimsValue(),
		},
	}

	_, info, err := ctx.Identities.GetByClaims(spec.Type, spec.Claims)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		info = nil
	} else if err != nil {
		return nil, err
	}

	return []newinteraction.Edge{
		&EdgeSelectIdentityEnd{RequestedIdentity: spec, ExistingIdentity: info},
	}, nil
}
