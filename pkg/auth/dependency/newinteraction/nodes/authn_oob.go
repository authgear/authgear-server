package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationOOB{})
}

type InputAuthenticationOOB interface {
	GetOOBOTP() string
}

type EdgeAuthenticationOOB struct {
	Stage         newinteraction.AuthenticationStage
	Authenticator *authenticator.Info
	Secret        string
}

func (e *EdgeAuthenticationOOB) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputAuthenticationOOB)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	info := e.Authenticator
	err := ctx.Authenticators.VerifySecret(info, map[string]string{
		authenticator.AuthenticatorStateOOBOTPSecret: e.Secret,
	}, input.GetOOBOTP())
	if err != nil {
		info = nil
	}

	return &NodeAuthenticationOOB{Stage: e.Stage, Authenticator: info}, nil
}

type NodeAuthenticationOOB struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeAuthenticationOOB) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOB) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeAuthenticationEnd{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
