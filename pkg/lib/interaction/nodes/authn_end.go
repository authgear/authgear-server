package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationEnd{})
}

type EdgeAuthenticationEnd struct {
	Stage                 authn.AuthenticationStage
	AuthenticationType    authn.AuthenticationType
	VerifiedAuthenticator *authenticator.Info
	RecoveryCode          *mfa.RecoveryCode
}

func (e *EdgeAuthenticationEnd) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	node := &NodeAuthenticationEnd{
		Stage:                 e.Stage,
		AuthenticationType:    e.AuthenticationType,
		VerifiedAuthenticator: e.VerifiedAuthenticator,
		RecoveryCode:          e.RecoveryCode,
	}

	return node, nil
}

type NodeAuthenticationEnd struct {
	Stage                 authn.AuthenticationStage `json:"stage"`
	AuthenticationType    authn.AuthenticationType  `json:"authentication_type"`
	VerifiedAuthenticator *authenticator.Info       `json:"verified_authenticator"`
	RecoveryCode          *mfa.RecoveryCode         `json:"recovery_code"`
}

func (n *NodeAuthenticationEnd) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationEnd) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return []interaction.Effect{
		interaction.EffectRun(func(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, nodeIndex int) error {
			if n.VerifiedAuthenticator == nil {
				return nil
			}
			return ctx.Authenticators.MarkOOBIdentityVerified(goCtx, n.VerifiedAuthenticator)
		}),
	}, nil
}

func (n *NodeAuthenticationEnd) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	err := n.IsFailure()
	if err != nil {
		return nil, err
	}

	return graph.Intent.DeriveEdgesForNode(goCtx, graph, n)
}

func (n *NodeAuthenticationEnd) FillDetails(err error) error {
	return errorutil.WithDetails(err, errorutil.Details{
		"AuthenticationType": apierrors.APIErrorDetail.Value(n.AuthenticationType),
	})
}

func (n *NodeAuthenticationEnd) IsFailure() (err error) {
	switch n.AuthenticationType {
	case authn.AuthenticationTypeNone:
		break
	case authn.AuthenticationTypePassword,
		authn.AuthenticationTypePasskey,
		authn.AuthenticationTypeTOTP,
		authn.AuthenticationTypeOOBOTPEmail,
		authn.AuthenticationTypeOOBOTPSMS:
		if n.VerifiedAuthenticator == nil {
			err = n.FillDetails(api.ErrInvalidCredentials)
			return
		}
	case authn.AuthenticationTypeRecoveryCode:
		if n.RecoveryCode == nil {
			err = n.FillDetails(api.ErrInvalidCredentials)
			return
		}
	case authn.AuthenticationTypeDeviceToken:
		break
	default:
		panic("interaction: unknown authentication type: " + n.AuthenticationType)
	}

	return
}
