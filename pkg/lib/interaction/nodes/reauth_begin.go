package nodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeReauthenticationBegin{})
}

type EdgeReauthenticationBegin struct{}

func (e *EdgeReauthenticationBegin) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	stage, _, err := e.getAuthenticators(goCtx, ctx, graph)
	if err != nil {
		// This is must be reported with panic.
		// If we return it as pain error,
		// /select_account is visited again, the error is generated again,
		// resulting in infinite redirection.
		if errors.Is(err, api.ErrNoAuthenticator) {
			panic(err)
		}
		return nil, err
	}
	return &NodeReauthenticationBegin{
		Stage: stage,
	}, nil
}

func (e *EdgeReauthenticationBegin) getAuthenticators(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) (stage authn.AuthenticationStage, authenticators []*authenticator.Info, err error) {
	ais, err := ctx.Authenticators.List(goCtx, graph.MustGetUserID())
	if err != nil {
		return
	}

	primary := [][]authenticator.Filter{
		// Primary password
		{
			authenticator.KeepKind(authenticator.KindPrimary),
			authenticator.KeepType(model.AuthenticatorTypePassword),
		},
		// Primary OOB email
		{
			authenticator.KeepKind(authenticator.KindPrimary),
			authenticator.KeepType(model.AuthenticatorTypeOOBEmail),
		},
		// Primary OOB SMS
		{
			authenticator.KeepKind(authenticator.KindPrimary),
			authenticator.KeepType(model.AuthenticatorTypeOOBSMS),
		},
	}

	secondary := [][]authenticator.Filter{
		// Secondary TOTP
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(model.AuthenticatorTypeTOTP),
		},
		// Secondary password
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(model.AuthenticatorTypePassword),
		},
		// Secondary OOB email
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(model.AuthenticatorTypeOOBEmail),
		},
		// Secondary OOB SMS
		{
			authenticator.KeepKind(authenticator.KindSecondary),
			authenticator.KeepType(model.AuthenticatorTypeOOBSMS),
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

	err = api.ErrNoAuthenticator
	return
}

type NodeReauthenticationBegin struct {
	Stage               authn.AuthenticationStage   `json:"stage"`
	Authenticators      []*authenticator.Info       `json:"-"`
	AuthenticatorConfig *config.AuthenticatorConfig `json:"-"`
}

func (n *NodeReauthenticationBegin) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	edge := &EdgeReauthenticationBegin{}
	stage, authenticators, err := edge.getAuthenticators(goCtx, ctx, graph)
	if err != nil {
		return err
	}

	if stage != n.Stage {
		panic(fmt.Errorf("interaction: the set of authenticators changed during reauthentication"))
	}

	n.Authenticators = authenticators
	n.AuthenticatorConfig = ctx.Config.Authenticator
	return nil
}

func (n *NodeReauthenticationBegin) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeReauthenticationBegin) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
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
		case model.AuthenticatorTypePassword:
			edges = append(edges, &EdgeAuthenticationPassword{
				Stage:          n.Stage,
				Authenticators: []*authenticator.Info{a},
			})
		case model.AuthenticatorTypeTOTP:
			edges = append(edges, &EdgeAuthenticationTOTP{
				Stage:          n.Stage,
				Authenticators: []*authenticator.Info{a},
			})
		case model.AuthenticatorTypeOOBEmail:
			edges = append(edges, &EdgeAuthenticationOOBTrigger{
				Stage:                n.Stage,
				Authenticators:       []*authenticator.Info{a},
				OOBAuthenticatorType: model.AuthenticatorTypeOOBEmail,
			})
		case model.AuthenticatorTypeOOBSMS:
			if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.Deprecated_IsWhatsappEnabled() {
				edges = append(edges, &EdgeAuthenticationWhatsappTrigger{
					Stage:          n.Stage,
					Authenticators: []*authenticator.Info{a},
				})
			}

			if n.AuthenticatorConfig.OOB.SMS.PhoneOTPMode.Deprecated_IsSMSEnabled() {
				edges = append(edges, &EdgeAuthenticationOOBTrigger{
					Stage:                n.Stage,
					Authenticators:       []*authenticator.Info{a},
					OOBAuthenticatorType: model.AuthenticatorTypeOOBSMS,
				})
			}
		}
	}

	return edges, nil
}
