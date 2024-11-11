package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationTOTP{})
}

type InputAuthenticationTOTP interface {
	GetTOTP() string
}

type EdgeAuthenticationTOTP struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationTOTP) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypeTOTP
}

func (e *EdgeAuthenticationTOTP) IsDefaultAuthenticator() bool {
	filtered := authenticator.ApplyFilters(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationTOTP) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationTOTP
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	spec := &authenticator.Spec{
		TOTP: &authenticator.TOTPSpec{
			Code: input.GetTOTP(),
		},
	}

	info, _, err := ctx.Authenticators.VerifyOneWithSpec(goCtx,
		graph.MustGetUserID(),
		model.AuthenticatorTypeTOTP,
		e.Authenticators,
		spec,
		&facade.VerifyOptions{
			AuthenticationDetails: facade.NewAuthenticationDetails(
				graph.MustGetUserID(),
				e.Stage,
				authn.AuthenticationTypeTOTP,
			),
		},
	)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationTOTP{Stage: e.Stage, Authenticator: info}, nil
}

type NodeAuthenticationTOTP struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeAuthenticationTOTP) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationTOTP) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationTOTP) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    authn.AuthenticationTypeTOTP,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}
