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
	interaction.RegisterNode(&NodeCreateAuthenticatorMagicLinkOTPSetup{})
}

type InputCreateAuthenticatorMagicLinkOTPSetup interface {
	GetMagicLinkOTPTarget() string
}

type EdgeCreateAuthenticatorMagicLinkOTPSetup struct {
	NewAuthenticatorID string
	Stage              authn.AuthenticationStage
	IsDefault          bool
}

type InputCreateAuthenticatorMagicLinkOTPSetupSelect interface {
	SetupPrimaryAuthenticatorMagicLinkOTP()
}

func (e *EdgeCreateAuthenticatorMagicLinkOTPSetup) IsDefaultAuthenticator() bool {
	return false
}

func (e *EdgeCreateAuthenticatorMagicLinkOTPSetup) AuthenticatorType() model.AuthenticatorType {
	// Currently only support send through email
	return model.AuthenticatorTypeOOBEmail
}

func (e *EdgeCreateAuthenticatorMagicLinkOTPSetup) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var userID string
	var input InputCreateAuthenticatorMagicLinkOTPSetup
	if e.Stage == authn.AuthenticationStagePrimary {
		panic("Magic link as primary authenticator is not yet supported")
	} else {
		if !interaction.Input(rawInput, &input) {
			return nil, interaction.ErrIncompatibleInput
		}
		userID = graph.MustGetUserID()
	}
	spec := &authenticator.Spec{
		UserID:    userID,
		IsDefault: e.IsDefault,
		Kind:      stageToAuthenticatorKind(e.Stage),
		Type:      e.AuthenticatorType(),
		OOBOTP: &authenticator.OOBOTPSpec{
			Email: input.GetMagicLinkOTPTarget(),
		},
		MagicLinkOTP: &authenticator.MagicLinkOTPSpec{
			Email: input.GetMagicLinkOTPTarget(),
		},
	}

	spec.RequiredToVerify = true
	_, isNewUser := graph.GetNewUserID()
	if isNewUser &&
		e.Stage == authn.AuthenticationStageSecondary &&
		ctx.Config.Authenticator.OOB.Email.SecondaryAllowUnverified {
		spec.RequiredToVerify = false
	}

	info, err := ctx.Authenticators.NewWithAuthenticatorID(e.NewAuthenticatorID, spec)
	if err != nil {
		return nil, err
	}

	var skipInput interface{ SkipVerification() bool }
	if interaction.Input(rawInput, &skipInput) && skipInput.SkipVerification() {
		// Admin skip verify MagicLink otp and create OOB authenticator directly
		return &NodeCreateAuthenticatorMagicLinkOTP{Stage: e.Stage, Authenticator: info}, nil
	}

	if !spec.RequiredToVerify {
		return &NodeCreateAuthenticatorMagicLinkOTP{Stage: e.Stage, Authenticator: info, DeferVerify: true}, nil
	}

	aStatus, err := ctx.Verification.GetAuthenticatorVerificationStatus(info)
	if err != nil {
		return nil, err
	}

	if aStatus == verification.AuthenticatorStatusVerified {
		return &NodeCreateAuthenticatorMagicLinkOTP{Stage: e.Stage, Authenticator: info}, nil
	}

	result, err := (&SendOOBCode{
		Context:              ctx,
		Stage:                e.Stage,
		IsAuthenticating:     false,
		AuthenticatorInfo:    info,
		IgnoreRatelimitError: true,
		OTPMode:              otp.OTPModeMagicLink,
	}).Do()
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorMagicLinkOTPSetup{
		Stage:         e.Stage,
		Authenticator: info,
		MagicLinkOTP:  result.Code,
		Target:        input.GetMagicLinkOTPTarget(),
		Channel:       result.Channel,
	}, nil
}

type NodeCreateAuthenticatorMagicLinkOTPSetup struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	MagicLinkOTP  string                    `json:"magic_link_otp"`
	Target        string                    `json:"target"`
	Channel       string                    `json:"channel"`
}

// GetMagicLinkOTP implements MagicLinkOTPNode.
func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetMagicLinkOTP() string {
	return n.MagicLinkOTP
}

// GetMagicLinkOTPTarget implements MagicLinkOTPNode.
func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetMagicLinkOTPTarget() string {
	return n.Target
}

// GetMagicLinkOTPChannel implements MagicLinkOTPNode.
func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetMagicLinkOTPChannel() string {
	return n.Channel
}

// GetMagicLinkOTPOOBType implements MagicLinkOTPNode.
func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetMagicLinkOTPOOBType() interaction.OOBType {
	switch n.Stage {
	case authn.AuthenticationStagePrimary:
		return interaction.OOBTypeSetupPrimary
	case authn.AuthenticationStageSecondary:
		return interaction.OOBTypeSetupSecondary
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}
}

func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeOOBResendCode{
			Stage:            n.Stage,
			IsAuthenticating: false,
			Authenticator:    n.Authenticator,
			OTPMode:          otp.OTPModeMagicLink,
		},
		&EdgeCreateAuthenticatorMagicLinkOTP{Stage: n.Stage, Authenticator: n.Authenticator},
	}
	return edges, nil
}
