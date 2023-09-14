package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/uiparam"
)

func init() {
	authflow.RegisterNode(&NodeOAuth{})
}

type NodeOAuthData struct {
	OAuthAuthorizationURL string `json:"oauth_authorization_url,omitempty"`
}

var _ authflow.Data = NodeOAuthData{}

func (NodeOAuthData) Data() {}

type NodeOAuth struct {
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	NewUserID   string        `json:"new_user_id,omitempty"`
	Alias       string        `json:"alias,omitempty"`
	State       string        `json:"state,omitempty"`
	RedirectURI string        `json:"redirect_uri,omitempty"`
}

var _ authflow.NodeSimple = &NodeOAuth{}
var _ authflow.InputReactor = &NodeOAuth{}
var _ authflow.DataOutputer = &NodeOAuth{}

func (*NodeOAuth) Kind() string {
	return "NodeOAuth"
}

func (n *NodeOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaTakeOAuthAuthorizationResponse{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputOAuth inputTakeOAuthAuthorizationResponse
	if authflow.AsInput(input, &inputOAuth) {
		if oauthError := inputOAuth.GetOAuthError(); oauthError != "" {
			errorDescription := inputOAuth.GetOAuthErrorDescription()
			errorURI := inputOAuth.GetOAuthErrorURI()

			return nil, sso.NewOAuthError(oauthError, errorDescription, errorURI)
		}

		oauthProvider := deps.OAuthProviderFactory.NewOAuthProvider(n.Alias)
		if oauthProvider == nil {
			return nil, api.ErrOAuthProviderNotFound
		}

		code := inputOAuth.GetOAuthAuthorizationCode()

		// TODO(authflow): support nonce in OAuth.
		emptyNonce := ""
		authInfo, err := oauthProvider.GetAuthInfo(
			sso.OAuthAuthorizationResponse{
				Code: code,
			},
			sso.GetAuthInfoParam{
				Nonce: emptyNonce,
			},
		)
		if err != nil {
			return nil, err
		}

		providerConfig := oauthProvider.Config()
		providerID := providerConfig.ProviderID()
		identitySpec := &identity.Spec{
			Type: model.IdentityTypeOAuth,
			OAuth: &identity.OAuthSpec{
				ProviderID:     providerID,
				SubjectID:      authInfo.ProviderUserID,
				RawProfile:     authInfo.ProviderRawProfile,
				StandardClaims: authInfo.StandardAttributes.ToClaims(),
			},
		}

		// signup
		if n.NewUserID != "" {
			info, err := newIdentityInfo(deps, n.NewUserID, identitySpec)
			if err != nil {
				return nil, err
			}

			return authflow.NewNodeSimple(&NodeDoCreateIdentity{
				Identity: info,
			}), nil
		}
		// Else login

		exactMatch, err := findExactOneIdentityInfo(deps, identitySpec)
		if err != nil {
			return nil, err
		}

		n, err := NewNodeDoUseIdentity(flows, &NodeDoUseIdentity{
			Identity: exactMatch,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeOAuth) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	oauthProvider := deps.OAuthProviderFactory.NewOAuthProvider(n.Alias)
	if oauthProvider == nil {
		return nil, api.ErrOAuthProviderNotFound
	}

	uiParam := uiparam.GetUIParam(ctx)

	param := sso.GetAuthURLParam{
		State:  n.State,
		Prompt: uiParam.Prompt,
	}

	authorizationURL, err := oauthProvider.GetAuthURL(param)
	if err != nil {
		return nil, err
	}

	return NodeOAuthData{
		OAuthAuthorizationURL: authorizationURL,
	}, nil
}
