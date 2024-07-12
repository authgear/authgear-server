package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeVerifyIdentityViaWhatsapp{})
}

type InputVerifyIdentityViaWhatsapp interface {
	SelectVerifyIdentityViaWhatsapp()
}

type EdgeVerifyIdentityViaWhatsapp struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (e *EdgeVerifyIdentityViaWhatsapp) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityViaWhatsapp
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	if err := ensurePhoneLoginIDIdentity(e.Identity); err != nil {
		panic(err)
	}

	phone := e.Identity.LoginID.LoginID

	result, err := NewSendWhatsappCode(ctx, otp.KindVerification, phone, false).Do()
	if err != nil {
		return nil, err
	}

	node := &NodeVerifyIdentityViaWhatsapp{
		Identity:          e.Identity,
		RequestedByUser:   e.RequestedByUser,
		WhatsappOTPLength: result.CodeLength,
		Phone:             phone,
	}
	return node, nil
}

type NodeVerifyIdentityViaWhatsapp struct {
	Identity          *identity.Info `json:"identity"`
	RequestedByUser   bool           `json:"requested_by_user"`
	WhatsappOTPLength int            `json:"whatsapp_otp_length"`
	Phone             string         `json:"phone"`
}

// GetWhatsappOTPLength implements WhatsappOTPNode.
func (n *NodeVerifyIdentityViaWhatsapp) GetWhatsappOTPLength() int {
	return n.WhatsappOTPLength
}

// GetPhone implements WhatsappOTPNode.
func (n *NodeVerifyIdentityViaWhatsapp) GetPhone() string {
	return n.Phone
}

// GetOTPKindFactory implements WhatsappOTPNode.
func (n *NodeVerifyIdentityViaWhatsapp) GetOTPKindFactory() otp.KindFactory {
	return otp.KindVerification
}

func (n *NodeVerifyIdentityViaWhatsapp) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeVerifyIdentityViaWhatsapp) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeVerifyIdentityViaWhatsapp) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeWhatsappOTPResendCode{
			Target:         n.Phone,
			OTPKindFactory: n.GetOTPKindFactory(),
		},
		&EdgeVerifyIdentityViaWhatsappCheckCode{Identity: n.Identity},
	}

	return edges, nil
}

type InputVerifyIdentityViaWhatsappCheckCode interface {
	GetWhatsappOTP() string
}

type EdgeVerifyIdentityViaWhatsappCheckCode struct {
	Identity *identity.Info `json:"identity"`
}

func (e *EdgeVerifyIdentityViaWhatsappCheckCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	if err := ensurePhoneLoginIDIdentity(e.Identity); err != nil {
		panic(err)
	}

	var input InputVerifyIdentityViaWhatsappCheckCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	phone := e.Identity.LoginID.LoginID
	userID := e.Identity.UserID
	code := input.GetWhatsappOTP()
	err := ctx.OTPCodeService.VerifyOTP(
		otp.KindVerification(ctx.Config, model.AuthenticatorOOBChannelWhatsapp),
		phone,
		code,
		&otp.VerifyOptions{
			UserID: userID,
		},
	)
	if err != nil {
		if apierrors.IsKind(err, otp.InvalidOTPCode) {
			return nil, verification.ErrInvalidVerificationCode
		}
		return nil, err
	}

	verifiedClaim := ctx.Verification.NewVerifiedClaim(
		e.Identity.UserID,
		string(model.ClaimPhoneNumber),
		phone,
	)
	return &NodeEnsureVerificationEnd{
		Identity:         e.Identity,
		NewVerifiedClaim: verifiedClaim,
	}, nil
}

func ensurePhoneLoginIDIdentity(info *identity.Info) error {
	if info.Type != model.IdentityTypeLoginID {
		return fmt.Errorf("interaction: expect login ID identity: %s", info.Type)
	}

	loginIDType := info.LoginID.LoginIDType
	if loginIDType != model.LoginIDKeyTypePhone {
		return fmt.Errorf("interaction: expect phone login id type: %s", info.Type)
	}

	return nil
}
