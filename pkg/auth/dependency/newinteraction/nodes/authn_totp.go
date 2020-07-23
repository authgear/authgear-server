package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

type InputAuthenticationTOTP interface {
	GetTOTP() string
}

type EdgeAuthenticationTOTP struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationTOTP) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputAuthenticationTOTP)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID: userID,
		Type:   authn.AuthenticatorTypeTOTP,
		Props:  map[string]interface{}{},
	}
	info, err := ctx.Authenticators.Authenticate(userID, *spec, nil, input.GetTOTP())
	if errors.Is(err, authenticator.ErrAuthenticatorNotFound) ||
		errors.Is(err, authenticator.ErrInvalidCredentials) {
		info = nil
	} else if err != nil {
		return nil, err
	}

	return &NodeAuthenticationTOTP{Stage: e.Stage, Authenticator: info}, nil
}

type NodeAuthenticationTOTP struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeAuthenticationTOTP) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationTOTP) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeAuthenticationEnd{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
