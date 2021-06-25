package nodes

import (
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeReauthenticationBegin{})
}

type EdgeReauthenticationBegin struct{}

func (e *EdgeReauthenticationBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	stage, _, err := e.getAuthenticators(ctx, graph)
	if err != nil {
		// This is must be reported with panic.
		// If we return it as pain error,
		// /select_account is visited again, the error is generated again,
		// resulting in infinite redirection.
		if errors.Is(err, interaction.ErrNoAuthenticator) {
			panic(err)
		}
		return nil, err
	}
	return &NodeReauthenticationBegin{
		Stage: stage,
	}, nil
}

func (e *EdgeReauthenticationBegin) getAuthenticators(ctx *interaction.Context, graph *interaction.Graph) (stage authn.AuthenticationStage, authenticators []*authenticator.Info, err error) {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return
	}

	primary := [][]authenticator.Filter{
		// Primary password
		{
			authenticator.KeepKind(authenticator.KindPrimary),
			authenticator.KeepType(authn.AuthenticatorTypePassword),
		},
		// Primary OOB email
		{
			authenticator.KeepKind(authenticator.KindPrimary),
			authenticator.KeepType(authn.AuthenticatorTypeOOBEmail),
		},
		// Primary OOB SMS
		{
			authenticator.KeepKind(authenticator.KindPrimary),
			authenticator.KeepType(authn.AuthenticatorTypeOOBSMS),
		},
	}

	secondary := [][]authenticator.Filter{
		// Secondary TOTP
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(authn.AuthenticatorTypeTOTP),
		},
		// Secondary password
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(authn.AuthenticatorTypePassword),
		},
		// Secondary OOB email
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(authn.AuthenticatorTypeOOBEmail),
		},
		// Secondary OOB SMS
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(authn.AuthenticatorTypeOOBSMS),
		},
	}

	var available []*authenticator.Info

	for _, filters := range primary {
		filtered := authenticator.ApplyFilters(
			ais,
			filters...,
		)
		if len(filtered) > 0 {
			available = append(available, filtered...)
		}
	}
	if len(available) > 0 {
		stage = authn.AuthenticationStagePrimary
		authenticators = available
		return
	}

	for _, filters := range secondary {
		filtered := authenticator.ApplyFilters(
			ais,
			filters...,
		)
		if len(filtered) > 0 {
			available = append(available, filtered...)
		}
	}
	if len(available) > 0 {
		stage = authn.AuthenticationStageSecondary
		authenticators = available
		return
	}

	err = interaction.ErrNoAuthenticator
	return
}

type NodeReauthenticationBegin struct {
	Stage          authn.AuthenticationStage `json:"stage"`
	Authenticators []*authenticator.Info     `json:"-"`
}

func (n *NodeReauthenticationBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	edge := &EdgeReauthenticationBegin{}
	stage, authenticators, err := edge.getAuthenticators(ctx, graph)
	if err != nil {
		return err
	}

	if stage != n.Stage {
		panic(fmt.Errorf("interaction: the set of authenticators changed during reauthentication"))
	}

	n.Authenticators = authenticators
	return nil
}

func (n *NodeReauthenticationBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeReauthenticationBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.GetAuthenticationEdges()
}

// GetAuthenticationStage implements AuthenticationBeginNode.
func (n *NodeReauthenticationBegin) GetAuthenticationStage() authn.AuthenticationStage {
	return n.Stage
}

// GetAuthenticationEdges implements AuthenticationBeginNode.
func (n *NodeReauthenticationBegin) GetAuthenticationEdges() ([]interaction.Edge, error) {
	var edges []interaction.Edge

	for _, a := range n.Authenticators {
		switch a.Type {
		case authn.AuthenticatorTypePassword:
			edges = append(edges, &EdgeAuthenticationPassword{
				Stage:          n.Stage,
				Authenticators: []*authenticator.Info{a},
			})
		case authn.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeAuthenticationTOTP{
				Stage:          n.Stage,
				Authenticators: []*authenticator.Info{a},
			})
		case authn.AuthenticatorTypeOOBEmail:
			edges = append(edges, &EdgeAuthenticationOOBTrigger{
				Stage:                n.Stage,
				Authenticators:       []*authenticator.Info{a},
				OOBAuthenticatorType: authn.AuthenticatorTypeOOBEmail,
			})
		case authn.AuthenticatorTypeOOBSMS:
			edges = append(edges, &EdgeAuthenticationOOBTrigger{
				Stage:                n.Stage,
				Authenticators:       []*authenticator.Info{a},
				OOBAuthenticatorType: authn.AuthenticatorTypeOOBSMS,
			})
		}
	}

	return edges, nil
}
