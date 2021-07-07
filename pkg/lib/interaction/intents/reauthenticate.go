package intents

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/interaction/nodes"
	"github.com/authgear/authgear-server/pkg/lib/session"
)

func init() {
	interaction.RegisterIntent(&IntentReauthenticate{})
}

type IntentReauthenticate struct {
	WebhookState string `json:"webhook_state,omitempty"`
	UserIDHint   string `json:"user_id_hint,omitempty"`
}

func (i *IntentReauthenticate) InstantiateRootNode(ctx *interaction.Context, graph *interaction.Graph) (interaction.Node, error) {
	edge := nodes.EdgeDoUseUser{UseUserID: i.UserIDHint}
	return edge.Instantiate(ctx, graph, i)
}

func (i *IntentReauthenticate) DeriveEdgesForNode(graph *interaction.Graph, node interaction.Node) ([]interaction.Edge, error) {
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
		return []interaction.Edge{
			&nodes.EdgeDoEnsureSession{
				CreateReason: session.CreateReasonReauthenticate,
				Mode:         nodes.EnsureSessionModeUpdateIfPossible,
			},
		}, nil
	case *nodes.NodeDoEnsureSession:
		// Intent is finished.
		return nil, nil
	default:
		panic(fmt.Errorf("interaction: unexpected node: %T", node))
	}
}

func (i *IntentReauthenticate) GetWebhookState() string {
	return i.WebhookState
}

var _ interaction.IntentWithWebhookState = &IntentAuthenticate{}
