package intents

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func init() {
	interaction.RegisterIntent(&IntentReauthenticate{})
}

type IntentReauthenticate struct {
	UserIDHint               string `json:"user_id_hint,omitempty"`
	SuppressIDPSessionCookie bool   `json:"suppress_idp_session_cookie"`
}

func (i *IntentReauthenticate) InstantiateRootNode(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserIDHint}
	return edge.Instantiate(goCtx, ctx, graph, i)
}

func (i *IntentReauthenticate) DeriveEdgesForNode(goCtx context.Context, graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
	switch node := node.(type) {
	case *nodes.NodeDoUseUser:
		return []interaction.Edge{
			&nodes.EdgeReauthenticationBegin{},
		}, nil
	case *nodes.NodeAuthenticationEnd:
		return []interaction.Edge{
			&nodes.EdgeDoUseAuthenticator{
				Stage:         node.Stage,
				Authenticator: node.VerifiedAuthenticator,
			},
		}, nil
	case *nodes.NodeDoUseAuthenticator:
		mode := nodes.EnsureSessionModeUpdateOrCreate
		if i.SuppressIDPSessionCookie {
			mode = nodes.EnsureSessionModeNoop
		}
		return []interaction.Edge{
			&nodes.EdgeDoEnsureSession{
				CreateReason: session.CreateReasonReauthenticate,
				Mode:         mode,
			},
		}, nil
	case *nodes.NodeDoEnsureSession:
		// Intent is finished.
		return nil, nil
	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}
