package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorWhatsappOTPSetup{})
}

type InputCreateAuthenticatorWhatsappOTPSetup interface {
	GetWhatsappPhone() string
}

type EdgeCreateAuthenticatorWhatsappOTPSetup struct {
	NewAuthenticatorID string
	Stage              authn.AuthenticationStage
	IsDefault          bool
}

type InputCreateAuthenticatorWhatsappOTPSetupSelect interface {
	SetupPrimaryAuthenticatorWhatsappOTP()
}

func (e *EdgeCreateAuthenticatorWhatsappOTPSetup) IsDefaultAuthenticator() bool {
	return false
}

func (e *EdgeCreateAuthenticatorWhatsappOTPSetup) AuthenticatorType() model.AuthenticatorType {
	return model.AuthenticatorTypeOOBSMS
}

func (e *EdgeCreateAuthenticatorWhatsappOTPSetup) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var userID string
	var phone string
	if e.Stage == authn.AuthenticationStagePrimary {
		var input InputCreateAuthenticatorWhatsappOTPSetupSelect
		matchedInput := interaction.Input(rawInput, &input)
		if !matchedInput && !interaction.IsAdminAPI(rawInput) {
			return nil, interaction.ErrIncompatibleInput
		}
		identityInfo := graph.MustGetUserLastIdentity()
		userID = identityInfo.UserID
		phone = identityInfo.LoginID.LoginID
	} else {
		var input InputCreateAuthenticatorWhatsappOTPSetup
		if !interaction.Input(rawInput, &input) {
			return nil, interaction.ErrIncompatibleInput
		}
		userID = graph.MustGetUserID()
		phone = input.GetWhatsappPhone()
	}

	spec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: e.IsDefault,
		Kind:      stageToAuthenticatorKind(e.Stage),
		Type:      e.AuthenticatorType(),
		OOBOTP: &authenticator.OOBOTPSpec{
			Phone: phone,
		},
	}

	info, err := ctx.Authenticators.NewWithAuthenticatorID(e.NewAuthenticatorID, spec)
	if err != nil {
		return nil, err
	}
	phone = info.OOBOTP.ToTarget()

	var skipInput interface{ SkipVerification() bool }
	if interaction.Input(rawInput, &skipInput) && skipInput.SkipVerification() {
		// Admin skip verify whatsapp otp and create OOB authenticator directly
		return &NodeCreateAuthenticatorOOB{Stage: e.Stage, Authenticator: info}, nil
	}

	// Skip checking whatsapp otp if the phone number is verified
	// Create OOB authenticator directly
	aStatus, err := ctx.Verification.GetAuthenticatorVerificationStatus(info)
	if err != nil {
		return nil, err
	}
	if aStatus == verification.AuthenticatorStatusVerified {
		return &NodeCreateAuthenticatorOOB{Stage: e.Stage, Authenticator: info}, nil
	}

	result, err := NewSendWhatsappCode(ctx, otp.KindOOBOTPCode, phone, false).Do()
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorWhatsappOTPSetup{
		Stage:             e.Stage,
		Authenticator:     info,
		WhatsappOTPLength: result.CodeLength,
		Phone:             phone,
	}, nil
}

type NodeCreateAuthenticatorWhatsappOTPSetup struct {
	Stage             authn.AuthenticationStage `json:"stage"`
	Authenticator     *authenticator.Info       `json:"authenticator"`
	WhatsappOTPLength int                       `json:"whatsapp_otp_length"`
	Phone             string                    `json:"phone"`
}

// GetWhatsappOTPLength implements WhatsappOTPNode.
func (n *NodeCreateAuthenticatorWhatsappOTPSetup) GetWhatsappOTPLength() int {
	return n.WhatsappOTPLength
}

// GetPhone implements WhatsappOTPNode.
func (n *NodeCreateAuthenticatorWhatsappOTPSetup) GetPhone() string {
	return n.Phone
}

// GetOTPKindFactory implements WhatsappOTPNode.
func (n *NodeCreateAuthenticatorWhatsappOTPSetup) GetOTPKindFactory() otp.KindFactory {
	return otp.KindOOBOTPCode
}

// GetCreateAuthenticatorStage implements CreateAuthenticatorPhoneOTPNode
func (n *NodeCreateAuthenticatorWhatsappOTPSetup) GetCreateAuthenticatorStage() authn.AuthenticationStage {
	return n.Stage
}

// GetSelectedPhoneNumberForPhoneOTP implements CreateAuthenticatorPhoneOTPNode
func (n *NodeCreateAuthenticatorWhatsappOTPSetup) GetSelectedPhoneNumberForPhoneOTP() string {
	return n.Phone
}

func (n *NodeCreateAuthenticatorWhatsappOTPSetup) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorWhatsappOTPSetup) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorWhatsappOTPSetup) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeWhatsappOTPResendCode{
			Target:         n.Phone,
			OTPKindFactory: n.GetOTPKindFactory(),
		},
		&EdgeCreateAuthenticatorWhatsappOTP{Stage: n.Stage, Authenticator: n.Authenticator},
	}
	return edges, nil
}
