package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationPasskey{})
}

type InputAuthenticationPasskey interface {
	GetAssertionResponse() []byte
}

type EdgeAuthenticationPasskey struct {
	Stage          authn.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationPasskey) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypePasskey
}

func (e *EdgeAuthenticationPasskey) IsDefaultAuthenticator() bool {
	filtered := authenticator.ApplyFilters(e.Authenticators, authenticator.KeepDefault)
	return len(filtered) > 0
}

func (e *EdgeAuthenticationPasskey) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var stageInput InputAuthenticationStage
	if !interaction.Input(rawInput, &stageInput) {
		return nil, interaction.ErrIncompatibleInput
	}
	stage := stageInput.GetAuthenticationStage()
	if stage != e.Stage {
		return nil, interaction.ErrIncompatibleInput
	}

	var input InputAuthenticationPasskey
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	assertionResponse := input.GetAssertionResponse()
	spec := &authenticator.Spec{
		Type: model.AuthenticatorTypePasskey,
		Passkey: &authenticator.PasskeySpec{
			AssertionResponse: assertionResponse,
		},
	}

	var requireUpdate bool
	var info *authenticator.Info
	for _, a := range e.Authenticators {
		b, err := ctx.Authenticators.VerifyWithSpec(a, spec)
		if errors.Is(err, authenticator.ErrInvalidCredentials) {
			continue
		} else if err != nil {
			return nil, err
		} else {
			aa := a
			info = aa
			requireUpdate = b
			break
		}
	}

	return &NodeAuthenticationPasskey{
		Stage:         e.Stage,
		Spec:          spec,
		Authenticator: info,
		RequireUpdate: requireUpdate,
	}, nil
}

type NodeAuthenticationPasskey struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Spec          *authenticator.Spec       `json:"spec"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	RequireUpdate bool                      `json:"require_update"`
}

func (n *NodeAuthenticationPasskey) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationPasskey) GetEffects() ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectOnCommit(func(ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			assertionResponse := n.Spec.Passkey.AssertionResponse

			err := ctx.Passkey.ConsumeAssertionResponse(assertionResponse)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (n *NodeAuthenticationPasskey) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    authn.AuthenticationTypePasskey,
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}

func (n *NodeAuthenticationPasskey) GetRequireUpdateAuthenticator(stage authn.AuthenticationStage) (info *authenticator.Info, ok bool) {
	if n.RequireUpdate && n.Stage == stage {
		return n.Authenticator, true
	}
	return nil, false
}
