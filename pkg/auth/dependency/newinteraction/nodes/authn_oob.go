package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
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

	if e.Authenticator == nil {
		return &NodeAuthenticationOOB{Stage: e.Stage, Authenticator: nil}, nil
	}

	userID := graph.MustGetUserID()
	spec := &authenticator.Spec{
		UserID: userID,
		Type:   authn.AuthenticatorTypeOOB,
		Props:  map[string]interface{}{},
	}
	info, err := ctx.Authenticators.Authenticate(userID, *spec, &map[string]string{
		authenticator.AuthenticatorStateOOBOTPID:          e.Authenticator.ID,
		authenticator.AuthenticatorStateOOBOTPSecret:      e.Secret,
		authenticator.AuthenticatorStateOOBOTPChannelType: e.Authenticator.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string),
	}, input.GetOOBOTP())
	if errors.Is(err, authenticator.ErrAuthenticatorNotFound) ||
		errors.Is(err, authenticator.ErrInvalidCredentials) {
		info = nil
	} else if err != nil {
		return nil, err
	}

	return &NodeAuthenticationOOB{Stage: e.Stage, Authenticator: info}, nil
}

type NodeAuthenticationOOB struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
}

func (n *NodeAuthenticationOOB) Apply(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOB) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeAuthenticationEnd{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
