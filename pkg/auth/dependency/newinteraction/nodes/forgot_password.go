package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

func init() {
	newinteraction.RegisterNode(&NodeForgotPasswordBegin{})
	newinteraction.RegisterNode(&NodeForgotPasswordEnd{})
}

type NodeForgotPasswordBegin struct{}

func (n *NodeForgotPasswordBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeForgotPasswordBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{&EdgeForgotPasswordSelectLoginID{Configs: ctx.Config.Identity.LoginID.Keys}}, nil
}

type InputForgotPasswordSelectLoginID interface {
	GetLoginID() string
}

type EdgeForgotPasswordSelectLoginID struct {
	Configs []config.LoginIDKeyConfig
}

// GetIdentityCandidates implements IdentityCandidatesGetter.
func (e *EdgeForgotPasswordSelectLoginID) GetIdentityCandidates() []identity.Candidate {
	candidates := make([]identity.Candidate, len(e.Configs))
	for i, c := range e.Configs {
		conf := c
		candidates[i] = identity.NewLoginIDCandidate(&conf)
	}
	return candidates
}

func (e *EdgeForgotPasswordSelectLoginID) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputForgotPasswordSelectLoginID)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	loginID := input.GetLoginID()

	err := ctx.ForgotPassword.SendCode(loginID)
	if err != nil {
		return nil, err
	}

	return &NodeForgotPasswordEnd{LoginID: loginID}, nil
}

type NodeForgotPasswordEnd struct {
	LoginID string `json:"login_id"`
}

// GetLoginID implements ForgotPasswordSuccessNode.
func (n *NodeForgotPasswordEnd) GetLoginID() string {
	return n.LoginID
}

func (n *NodeForgotPasswordEnd) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeForgotPasswordEnd) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return nil, nil
}
