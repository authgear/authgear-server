package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/validation"
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
	var target string
	if e.Stage == authn.AuthenticationStagePrimary {
		var input InputCreateAuthenticatorMagicLinkOTPSetupSelect
		matchedInput := interaction.Input(rawInput, &input)
		if !matchedInput && !interaction.IsAdminAPI(rawInput) {
			return nil, interaction.ErrIncompatibleInput
		}
		identityInfo := graph.MustGetUserLastIdentity()
		target = identityInfo.LoginID.LoginID
	} else {
		var input InputCreateAuthenticatorMagicLinkOTPSetup
		if !interaction.Input(rawInput, &input) {
			return nil, interaction.ErrIncompatibleInput
		}
		target = input.GetMagicLinkOTPTarget()
	}

	// Validate target against channel
	validationCtx := &validation.Context{}
	err := validation.FormatEmail{AllowName: false}.CheckFormat(target)
	if err != nil {
		validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
	}

	err = validationCtx.Error("invalid target")
	if err != nil {
		return nil, err
	}

	var spec *authenticator.Spec
	var identityInfo *identity.Info
	if e.Stage == authn.AuthenticationStagePrimary {
		// Primary OOB authenticators must be bound to login ID identity
		identityInfo = graph.MustGetUserLastIdentity()
		if identityInfo.Type != model.IdentityTypeLoginID {
			panic("interaction: OOB authenticator identity must be login ID")
		}

		spec = &authenticator.Spec{
			UserID:    identityInfo.UserID,
			IsDefault: e.IsDefault,
			Kind:      stageToAuthenticatorKind(e.Stage),
			Type:      model.AuthenticatorTypeOOBEmail,
			OOBOTP:    &authenticator.OOBOTPSpec{},
		}
	} else {
		userID := graph.MustGetUserID()
		spec = &authenticator.Spec{
			UserID:    userID,
			IsDefault: e.IsDefault,
			Kind:      stageToAuthenticatorKind(e.Stage),
			Type:      model.AuthenticatorTypeOOBEmail,
			OOBOTP:    &authenticator.OOBOTPSpec{},
		}

		// Normalize the target.
		var err error
		target, err = ctx.LoginIDNormalizerFactory.NormalizerWithLoginIDType(model.LoginIDKeyTypeEmail).Normalize(target)
		if err != nil {
			return nil, err
		}
	}

	spec.OOBOTP.Email = target

	info, err := ctx.Authenticators.NewWithAuthenticatorID(e.NewAuthenticatorID, spec)
	if err != nil {
		return nil, err
	}

	var skipInput interface{ SkipVerification() bool }
	if interaction.Input(rawInput, &skipInput) && skipInput.SkipVerification() {
		// Admin skip verify MagicLink otp and create OOB authenticator directly
		return &NodeCreateAuthenticatorMagicLinkOTP{Stage: e.Stage, Authenticator: info}, nil
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
		Target:        target,
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
