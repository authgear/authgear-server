package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationPassword{})
}

type InputAuthenticationPassword interface {
	GetPassword() string
}

type EdgeAuthenticationPassword struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationPassword) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputAuthenticationPassword)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID: userID,
		Type:   authn.AuthenticatorTypePassword,
		Props:  map[string]interface{}{},
	}
	info, err := ctx.Authenticators.Authenticate(spec, nil, input.GetPassword())
	if errors.Is(err, authenticator.ErrAuthenticatorNotFound) ||
		errors.Is(err, authenticator.ErrInvalidCredentials) {
		info = nil
	} else if err != nil {
		return nil, err
	}

	return &NodeAuthenticationPassword{Stage: e.Stage, Authenticator: info}, nil
}

type NodeAuthenticationPassword struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeAuthenticationPassword) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationPassword) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
