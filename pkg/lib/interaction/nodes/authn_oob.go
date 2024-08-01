package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeAuthenticationOOB{})
}

type InputAuthenticationOOB interface {
	GetOOBOTP() string
}

type EdgeAuthenticationOOB struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
	Secret        string
}

func (e *EdgeAuthenticationOOB) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputAuthenticationOOB
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	info := e.Authenticator
	_, err := ctx.Authenticators.VerifyWithSpec(info, &authenticator.Spec{
		OOBOTP: &authenticator.OOBOTPSpec{
			Code: input.GetOOBOTP(),
		},
	}, &facade.VerifyOptions{
		Form: otp.FormCode,
		AuthenticationDetails: facade.NewAuthenticationDetails(
			info.UserID,
			e.Stage,
			deriveNodeAuthenticationOOBAuthenticationType(info.Type),
		),
	})
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationOOB{Stage: e.Stage, Authenticator: info, AuthenticatorType: e.Authenticator.Type}, nil
}

type NodeAuthenticationOOB struct {
	Stage             authn.AuthenticationStage `json:"stage"`
	AuthenticatorType model.AuthenticatorType   `json:"authenticator_type"`
	Authenticator     *authenticator.Info       `json:"authenticator"`
}

func (n *NodeAuthenticationOOB) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOB) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeAuthenticationOOB) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeAuthenticationEnd{
			Stage:                 n.Stage,
			AuthenticationType:    deriveNodeAuthenticationOOBAuthenticationType(n.AuthenticatorType),
			VerifiedAuthenticator: n.Authenticator,
		},
	}, nil
}

func deriveNodeAuthenticationOOBAuthenticationType(authenticatorType model.AuthenticatorType) authn.AuthenticationType {
	var typ authn.AuthenticationType
	switch authenticatorType {
	case model.AuthenticatorTypeOOBEmail:
		typ = authn.AuthenticationTypeOOBOTPEmail
	case model.AuthenticatorTypeOOBSMS:
		typ = authn.AuthenticationTypeOOBOTPSMS
	default:
		panic(fmt.Errorf("interaction: unexpected authenticator type: %v", authenticatorType))
	}
	return typ
}
