package nodes

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorWhatsappOTP{})
}

type InputCreateAuthenticatorWhatsappOTP interface {
	GetWhatsappOTP() string
}

type EdgeCreateAuthenticatorWhatsappOTP struct {
	Stage         authn.AuthenticationStage
	Authenticator *authenticator.Info
}

func (e *EdgeCreateAuthenticatorWhatsappOTP) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputCreateAuthenticatorWhatsappOTP
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}
	phone := e.Authenticator.OOBOTP.Phone
	userID := e.Authenticator.UserID
	code := input.GetWhatsappOTP()
	err := ctx.OTPCodeService.VerifyOTP(
		otp.KindOOBOTP(ctx.Config, model.AuthenticatorOOBChannelWhatsapp),
		phone,
		code,
		&otp.VerifyOptions{
			UserID: userID,
		},
	)
	if err != nil {
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, errorutil.WithDetails(api.ErrInvalidCredentials, errorutil.Details{
				"AuthenticationType": apierrors.APIErrorDetail.Value(e.Authenticator.Type),
			})
		}
		return nil, err
	}

	return &NodeCreateAuthenticatorWhatsappOTP{Stage: e.Stage, Authenticator: e.Authenticator}, nil
}

type NodeCreateAuthenticatorWhatsappOTP struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
}

func (n *NodeCreateAuthenticatorWhatsappOTP) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorWhatsappOTP) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorWhatsappOTP) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeCreateAuthenticatorEnd{
			Stage:          n.Stage,
			Authenticators: []*authenticator.Info{n.Authenticator},
		},
	}, nil
}
