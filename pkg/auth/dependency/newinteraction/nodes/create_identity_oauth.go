package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateIdentityOAuth{})
}

type InputCreateIdentityOAuth interface {
	GetUserInfo() sso.AuthInfo
}

type EdgeCreateIdentityOAuth struct {
	RequestedIdentity *identity.Spec
	Config            config.OAuthSSOProviderConfig
}

func (e *EdgeCreateIdentityOAuth) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	var claims map[string]interface{}
	if e.RequestedIdentity != nil {
		claims = e.RequestedIdentity.Claims
	} else {
		input, ok := rawInput.(InputCreateIdentityOAuth)
		if !ok {
			return nil, newinteraction.ErrIncompatibleInput
		}

		userInfo := input.GetUserInfo()
		providerID := userInfo.ProviderConfig.ProviderID()
		if !e.Config.ProviderID().Equal(&providerID) {
			return nil, newinteraction.ErrIncompatibleInput
		}

		claims = map[string]interface{}{
			identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
			identity.IdentityClaimOAuthSubjectID:    userInfo.ProviderUserInfo.ID,
			identity.IdentityClaimOAuthProfile:      userInfo.ProviderRawProfile,
			identity.IdentityClaimOAuthClaims:       userInfo.ProviderUserInfo.ClaimsValue(),
		}
	}

	newIdentity, err := ctx.Identities.New(graph.MustGetUserID(), &identity.Spec{
		Type:   authn.IdentityTypeOAuth,
		Claims: claims,
	})
	if err != nil {
		return nil, err
	}

	return &NodeCreateIdentityOAuth{Identity: newIdentity}, nil
}

type NodeCreateIdentityOAuth struct {
	Identity *identity.Info `json:"identity"`
}

func (n *NodeCreateIdentityOAuth) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateIdentityOAuth) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{&EdgeCreateIdentityEnd{NewIdentity: n.Identity}}, nil
}
