package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/sso"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type InputSelectIdentityOAuth interface {
	GetUserInfo() sso.AuthInfo
}

type EdgeSelectIdentityOAuth struct {
	Config config.OAuthSSOProviderConfig
}

func (e *EdgeSelectIdentityOAuth) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputSelectIdentityOAuth)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	return &NodeSelectIdentityOAuth{
		UserInfo: input.GetUserInfo(),
	}, nil
}

type NodeSelectIdentityOAuth struct {
	UserInfo sso.AuthInfo `json:"auth_info"`
}

func (n *NodeSelectIdentityOAuth) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeSelectIdentityOAuth) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	providerID := n.UserInfo.ProviderConfig.ProviderID()
	claims := map[string]interface{}{
		identity.IdentityClaimOAuthProviderKeys: providerID.Claims(),
		identity.IdentityClaimOAuthSubjectID:    n.UserInfo.ProviderUserInfo.ID,
		identity.IdentityClaimOAuthProfile:      n.UserInfo.ProviderRawProfile,
		identity.IdentityClaimOAuthClaims:       n.UserInfo.ProviderUserInfo.ClaimsValue(),
	}

	_, i, err := ctx.Identities.GetByClaims(authn.IdentityTypeOAuth, claims)
	if errors.Is(err, identity.ErrIdentityNotFound) {
		// TODO: create new OAuth identity
		i = nil

	} else if err != nil {
		return nil, err
	}

	return []newinteraction.Edge{
		&EdgeSelectIdentityEnd{Identity: i},
	}, nil
}
