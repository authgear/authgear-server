package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeVerifyIdentityViaWhatsapp{})
	interaction.RegisterNode(&NodeVerifyIdentityViaWhatsappFallbackSMS{})
}

type EdgeVerifyIdentityViaWhatsapp struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (e *EdgeVerifyIdentityViaWhatsapp) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	if err := ensurePhoneLoginIDIdentity(e.Identity); err != nil {
		panic(err)
	}

	phone := e.Identity.Claims[identity.IdentityClaimLoginIDValue].(string)

	code, err := ctx.WhatsappCodeProvider.CreateCode(phone, string(ctx.Config.ID), ctx.WebSessionID)
	if err != nil {
		return nil, err
	}

	node := &NodeVerifyIdentityViaWhatsapp{
		Identity:        e.Identity,
		RequestedByUser: e.RequestedByUser,
		WhatsappOTP:     code.Code,
		Phone:           phone,
		PhoneOTPMode:    ctx.Config.Authenticator.OOB.SMS.PhoneOTPMode,
	}
	return node, nil
}

type NodeVerifyIdentityViaWhatsapp struct {
	Identity        *identity.Info                   `json:"identity"`
	RequestedByUser bool                             `json:"requested_by_user"`
	WhatsappOTP     string                           `json:"whatsapp_otp"`
	Phone           string                           `json:"phone"`
	PhoneOTPMode    config.AuthenticatorPhoneOTPMode `json:"phone_otp_mode"`
}

// GetPhoneOTPMode implements WhatsappOTPNode.
func (n *NodeVerifyIdentityViaWhatsapp) GetPhoneOTPMode() config.AuthenticatorPhoneOTPMode {
	return n.PhoneOTPMode
}

// GetWhatsappOTP implements WhatsappOTPNode.
func (n *NodeVerifyIdentityViaWhatsapp) GetWhatsappOTP() string {
	return n.WhatsappOTP
}

// GetPhone implements WhatsappOTPNode.
func (n *NodeVerifyIdentityViaWhatsapp) GetPhone() string {
	return n.Phone
}

func (n *NodeVerifyIdentityViaWhatsapp) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeVerifyIdentityViaWhatsapp) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeVerifyIdentityViaWhatsapp) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeVerifyIdentityViaWhatsappCheckCode{Identity: n.Identity},
	}

	if n.PhoneOTPMode == config.AuthenticatorPhoneOTPModeWhatsappSMS {
		edges = append(edges, &EdgeVerifyIdentityViaWhatsappFallbackSMS{
			Identity:        n.Identity,
			RequestedByUser: n.RequestedByUser,
		})
	}

	return edges, nil
}

type InputVerifyIdentityViaWhatsappCheckCode interface {
	VerifyWhatsappOTP()
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

	phone := e.Identity.Claims[identity.IdentityClaimLoginIDValue].(string)
	_, err := ctx.WhatsappCodeProvider.VerifyCode(phone, ctx.WebSessionID, true)
	if err != nil {
		return nil, err
	}

	verifiedClaim := ctx.Verification.NewVerifiedClaim(
		e.Identity.UserID,
		identity.StandardClaimPhoneNumber,
		phone,
	)
	return &NodeEnsureVerificationEnd{
		Identity:         e.Identity,
		NewVerifiedClaim: verifiedClaim,
	}, nil
}

type InputVerifyIdentityViaWhatsappFallbackSMS interface {
	FallbackSMS()
}

type EdgeVerifyIdentityViaWhatsappFallbackSMS struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (e *EdgeVerifyIdentityViaWhatsappFallbackSMS) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityViaWhatsappFallbackSMS
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	return &NodeVerifyIdentityViaWhatsappFallbackSMS{
		Identity:        e.Identity,
		RequestedByUser: e.RequestedByUser,
	}, nil
}

type NodeVerifyIdentityViaWhatsappFallbackSMS struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (n *NodeVerifyIdentityViaWhatsappFallbackSMS) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeVerifyIdentityViaWhatsappFallbackSMS) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeVerifyIdentityViaWhatsappFallbackSMS) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeVerifyIdentity{
			Identity:        n.Identity,
			RequestedByUser: n.RequestedByUser,
		},
	}, nil
}

func ensurePhoneLoginIDIdentity(info *identity.Info) error {
	if info.Type != model.IdentityTypeLoginID {
		return fmt.Errorf("interaction: expect login ID identity: %s", info.Type)
	}

	loginIDType := info.Claims[identity.IdentityClaimLoginIDType].(string)
	if loginIDType != string(config.LoginIDKeyTypePhone) {
		return fmt.Errorf("interaction: expect phone login id type: %s", info.Type)
	}

	return nil
}
