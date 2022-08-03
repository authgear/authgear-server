package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorTOTPSetup{})
}

type InputCreateAuthenticatorTOTPSetup interface {
	SetupTOTP()
}

type EdgeCreateAuthenticatorTOTPSetup struct {
	NewAuthenticatorID string
	Stage              authn.AuthenticationStage
	IsDefault          bool
}

func (e *EdgeCreateAuthenticatorTOTPSetup) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypeTOTP
}

func (e *EdgeCreateAuthenticatorTOTPSetup) IsDefaultAuthenticator() bool {
	return false
}

func (e *EdgeCreateAuthenticatorTOTPSetup) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputCreateAuthenticatorTOTPSetup
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: e.IsDefault,
		Kind:      stageToAuthenticatorKind(e.Stage),
		Type:      model.AuthenticatorTypeTOTP,
		Claims: map[authenticator.ClaimKey]interface{}{
			// The display name will be filled in in a later node.
			authenticator.AuthenticatorClaimTOTPDisplayName: "",
		},
	}

	info, err := ctx.Authenticators.NewWithAuthenticatorID(e.NewAuthenticatorID, spec)
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorTOTPSetup{Stage: e.Stage, Authenticator: info}, nil
}

type NodeCreateAuthenticatorTOTPSetup struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorTOTPSetup) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorTOTPSetup) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorTOTPSetup) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
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
