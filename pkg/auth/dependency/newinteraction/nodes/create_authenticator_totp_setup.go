package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorTOTPSetup{})
}

type EdgeCreateAuthenticatorTOTPSetup struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeCreateAuthenticatorTOTPSetup) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	// This edge does not take any input so it always instantiate.
	// The reason bebind this change is because the previous node should not beware of NodeCreateAuthenticatorTOTPSetup is after it.
	// So no handler can implement the input of EdgeCreateAuthenticatorTOTPSetup.
	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID: userID,
		Tag:    stageToAuthenticatorTag(e.Stage),
		Type:   authn.AuthenticatorTypeTOTP,
		Props: map[string]interface{}{
			// The display name will be filled in in a later node.
			authenticator.AuthenticatorPropTOTPDisplayName: "",
		},
	}

	info, err := ctx.Authenticators.New(spec, "")
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorTOTPSetup{Stage: e.Stage, Authenticator: info}, nil
}

type NodeCreateAuthenticatorTOTPSetup struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorTOTPSetup) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorTOTPSetup) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorTOTPSetup) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeCreateAuthenticatorTOTP{
			Stage:         n.Stage,
			Authenticator: n.Authenticator,
		},
	}, nil
}

// GetTOTPAuthenticator implements SetupTOTPNode.
func (n *NodeCreateAuthenticatorTOTPSetup) GetTOTPAuthenticator() *authenticator.Info {
	return n.Authenticator
}
