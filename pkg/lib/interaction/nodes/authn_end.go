package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/mfa"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
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

func (e *EdgeAuthenticationEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, input interface{}) (interaction.Node, error) {
	node := &NodeAuthenticationEnd{
		Stage:                 e.Stage,
		AuthenticationType:    e.AuthenticationType,
		VerifiedAuthenticator: e.VerifiedAuthenticator,
		RecoveryCode:          e.RecoveryCode,
	}

	if err := node.IsFailure(); err != nil {
		userID := graph.MustGetUserID()
		user, err := ctx.Users.Get(userID, accesscontrol.EmptyRole)
		if err != nil {
			return nil, err
		}
		err = ctx.Events.DispatchEvent(&nonblocking.AuthenticationFailedEventPayload{
			User:                *user,
			AuthenticationStage: string(e.Stage),
			AuthenticationType:  string(e.AuthenticationType),
		})
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

type NodeAuthenticationEnd struct {
	Stage                 authn.AuthenticationStage `json:"stage"`
	AuthenticationType    authn.AuthenticationType  `json:"authentication_type"`
	VerifiedAuthenticator *authenticator.Info       `json:"verified_authenticator"`
	RecoveryCode          *mfa.RecoveryCode         `json:"recovery_code"`
}

func (n *NodeAuthenticationEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	err := n.IsFailure()
	if err != nil {
		return nil, err
	}

	return graph.Intent.DeriveEdgesForNode(graph, n)
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
		authn.AuthenticationTypeTOTP,
		authn.AuthenticationTypeOOBOTPEmail,
		authn.AuthenticationTypeOOBOTPSMS:
		if n.VerifiedAuthenticator == nil {
			err = n.FillDetails(interaction.ErrInvalidCredentials)
			return
		}
	case authn.AuthenticationTypeRecoveryCode:
		if n.RecoveryCode == nil {
			err = n.FillDetails(interaction.ErrInvalidCredentials)
			return
		}
	case authn.AuthenticationTypeDeviceToken:
		break
	default:
		panic("interaction: unknown authentication type: " + n.AuthenticationType)
	}

	return
}
