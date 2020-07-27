package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorTOTPSetup{})
}

type InputCreateAuthenticatorTOTPSetup interface {
	SetupTOTP() bool
}

type EdgeCreateAuthenticatorTOTPSetup struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeCreateAuthenticatorTOTPSetup) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	_, ok := rawInput.(InputCreateAuthenticatorTOTPSetup)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID: userID,
		Type:   authn.AuthenticatorTypeTOTP,
		Props:  map[string]interface{}{},
	}

	infos, err := ctx.Authenticators.New(spec.UserID, *spec, "")
	if err != nil {
		return nil, err
	}

	if len(infos) != 1 {
		panic("interaction: unexpected number of new TOTP authenticators")
	}

	return &NodeCreateAuthenticatorTOTPSetup{Stage: e.Stage, Authenticator: infos[0]}, nil
}

type NodeCreateAuthenticatorTOTPSetup struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorTOTPSetup) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorTOTPSetup) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeCreateAuthenticatorTOTP{
			Stage:         n.Stage,
			Authenticator: n.Authenticator,
		},
	}, nil
}
