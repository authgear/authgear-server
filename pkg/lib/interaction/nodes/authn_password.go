package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationPassword{})
}

type InputAuthenticationPassword interface {
	GetPassword() string
}

type EdgeAuthenticationPassword struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationPassword) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypePassword
}

func (e *EdgeAuthenticationPassword) IsDefaultAuthenticator() bool {
	filtered := authenticator.ApplyFilters(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationPassword) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	// We first check the stage so that if password + additional password is used,
	// we do not advance too far.
	// That is, we do not check the given primary password against secondary password and report error.
	var stageInput InputAuthenticationStage
	if !interaction.Input(rawInput, &stageInput) {
		return nil, interaction.ErrIncompatibleInput
	}
	stage := stageInput.GetAuthenticationStage()
	if stage != e.Stage {
		return nil, interaction.ErrIncompatibleInput
	}

	var passwordInput InputAuthenticationPassword
	if !interaction.Input(rawInput, &passwordInput) {
		return nil, interaction.ErrIncompatibleInput
	}

	inputPassword := passwordInput.GetPassword()
	spec := &authenticator.Spec{
		Password: &authenticator.PasswordSpec{
			PlainPassword: inputPassword,
		},
	}

	info, requireUpdate, err := ctx.Authenticators.VerifyOneWithSpec(
		graph.MustGetUserID(),
		model.AuthenticatorTypePassword,
		e.Authenticators,
		spec,
		&facade.VerifyOptions{
			AuthenticationDetails: facade.NewAuthenticationDetails(
				graph.MustGetUserID(),
				e.Stage,
				authn.AuthenticationTypePassword,
			),
		},
	)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationPassword{Stage: e.Stage, Authenticator: info, RequireUpdate: requireUpdate}, nil
}

type NodeAuthenticationPassword struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	RequireUpdate bool                      `json:"require_update"`
}

func (n *NodeAuthenticationPassword) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationPassword) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationPassword) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    authn.AuthenticationTypePassword,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}

func (n *NodeAuthenticationPassword) GetRequireUpdateAuthenticator(stage authn.AuthenticationStage) (info *authenticator.Info, ok bool) {
	if n.RequireUpdate && n.Stage == stage {
		return n.Authenticator, true
	}
	return nil, false
}
