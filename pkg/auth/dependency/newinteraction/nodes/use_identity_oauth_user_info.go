package nodes

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

func init() {
	newinteraction.RegisterNode(&NodeUseIdentityOAuthUserInfo{})
}

type InputUseIdentityOAuthUserInfo interface {
	GetProviderAlias() string
	GetNonceSource() *http.Cookie
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

func (e *EdgeUseIdentityOAuthUserInfo) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputUseIdentityOAuthUserInfo)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	alias := input.GetProviderAlias()
	nonceSource := input.GetNonceSource()
	code := input.GetCode()
	state := ctx.WebStateID
	scope := input.GetScope()
	oauthError := input.GetError()
	errorDescription := input.GetErrorDescription()
	hashedNonce := e.HashedNonce

	if e.Config.Alias != alias {
		return nil, fmt.Errorf("interaction: unexpected provider alias %s != %s", e.Config.Alias, alias)
	}

	oauthProvider := ctx.OAuthProviderFactory.NewOAuthProvider(alias)
	if oauthProvider == nil {
		return nil, newinteraction.ErrOAuthProviderNotFound
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

func (n *NodeUseIdentityOAuthUserInfo) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseIdentityOAuthUserInfo) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeUseIdentityOAuthUserInfo) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	if n.IsCreating {
		return []newinteraction.Edge{&EdgeCreateIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
	}
	return []newinteraction.Edge{&EdgeSelectIdentityEnd{IdentitySpec: n.IdentitySpec}}, nil
}
