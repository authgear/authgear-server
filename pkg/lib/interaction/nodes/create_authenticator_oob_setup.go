package nodes

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorOOBSetup{})
}

type InputCreateAuthenticatorOOBSetup interface {
	GetOOBChannel() model.AuthenticatorOOBChannel
	GetOOBTarget() string
}

type InputCreateAuthenticatorOOBSetupSelect interface {
	SetupPrimaryAuthenticatorOOB()
}

type EdgeCreateAuthenticatorOOBSetup struct {
	NewAuthenticatorID string
	Stage              authn.AuthenticationStage
	IsDefault          bool

	OOBAuthenticatorType model.AuthenticatorType
}

func (e *EdgeCreateAuthenticatorOOBSetup) AuthenticatorType() model.AuthenticatorType {
	return e.OOBAuthenticatorType
}

func (e *EdgeCreateAuthenticatorOOBSetup) IsDefaultAuthenticator() bool {
	return false
}

// nolint: gocognit
func (e *EdgeCreateAuthenticatorOOBSetup) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var target string
	var channel model.AuthenticatorOOBChannel
	if e.Stage == authn.AuthenticationStagePrimary {
		var input InputCreateAuthenticatorOOBSetupSelect
		matchedInput := interaction.Input(rawInput, &input)
		if !matchedInput && !interaction.IsAdminAPI(rawInput) {
			return nil, interaction.ErrIncompatibleInput
		}
		identityInfo := graph.MustGetUserLastIdentity()
		target = identityInfo.LoginID.LoginID
		loginIDType := identityInfo.LoginID.LoginIDType
		switch loginIDType {
		case model.LoginIDKeyTypePhone:
			channel = model.AuthenticatorOOBChannelSMS
		case model.LoginIDKeyTypeEmail:
			channel = model.AuthenticatorOOBChannelEmail
		default:
			panic(fmt.Sprintf("interaction: unexpected login id type: %s", loginIDType))
		}
	} else {
		var input InputCreateAuthenticatorOOBSetup
		if !interaction.Input(rawInput, &input) {
			return nil, interaction.ErrIncompatibleInput
		}
		channel = input.GetOOBChannel()
		if channel == "" {
			return nil, interaction.ErrIncompatibleInput
		}
		target = input.GetOOBTarget()
	}

	var spec *authenticator.Spec
	var identityInfo *identity.Info
	var oobAuthenticatorType model.AuthenticatorType
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
			OOBOTP:    &authenticator.OOBOTPSpec{},
		}

		// Ignore given OOB target, use channel & target inferred from identity
		loginIDKey := identityInfo.LoginID.LoginIDKey
		for _, t := range ctx.Config.Identity.LoginID.Keys {
			if t.Key != loginIDKey {
				continue
			}
			switch t.Type {
			case model.LoginIDKeyTypeEmail:
				channel = model.AuthenticatorOOBChannelEmail
				oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
			case model.LoginIDKeyTypePhone:
				channel = model.AuthenticatorOOBChannelSMS
				oobAuthenticatorType = model.AuthenticatorTypeOOBSMS
			default:
				panic("interaction: creating OOB authenticator for invalid login ID type")
			}
			break
		}
		if oobAuthenticatorType == "" {
			panic("interaction: login ID not found for creating OOB authenticator")
		}
		target = identityInfo.LoginID.LoginID

	} else {
		userID := graph.MustGetUserID()
		spec = &authenticator.Spec{
			UserID:    userID,
			IsDefault: e.IsDefault,
			Kind:      stageToAuthenticatorKind(e.Stage),
			OOBOTP:    &authenticator.OOBOTPSpec{},
		}

		switch channel {
		case model.AuthenticatorOOBChannelEmail:
			oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
		case model.AuthenticatorOOBChannelSMS:
			oobAuthenticatorType = model.AuthenticatorTypeOOBSMS
		default:
			panic("interaction: creating OOB authenticator for invalid channel")
		}
	}

	spec.Type = oobAuthenticatorType
	switch channel {
	case model.AuthenticatorOOBChannelSMS:
		spec.OOBOTP.Phone = target
	case model.AuthenticatorOOBChannelEmail:
		spec.OOBOTP.Email = target
	}

	info, err := ctx.Authenticators.NewWithAuthenticatorID(goCtx, e.NewAuthenticatorID, spec)
	if err != nil {
		return nil, err
	}
	target = info.OOBOTP.ToTarget()

	var skipInput interface{ SkipVerification() bool }
	if interaction.Input(rawInput, &skipInput) && skipInput.SkipVerification() {
		// Skip verification of OOB target
		return &NodeCreateAuthenticatorOOB{Stage: e.Stage, Authenticator: info}, nil
	}

	aStatus, err := ctx.Verification.GetAuthenticatorVerificationStatus(goCtx, info)
	if err != nil {
		return nil, err
	}

	if aStatus == verification.AuthenticatorStatusVerified {
		return &NodeCreateAuthenticatorOOB{Stage: e.Stage, Authenticator: info}, nil
	}

	result, err := (&SendOOBCode{
		Context:              ctx,
		Stage:                e.Stage,
		IsAuthenticating:     false,
		AuthenticatorInfo:    info,
		IgnoreRatelimitError: true,
		OTPForm:              otp.FormCode,
	}).Do(goCtx)
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorOOBSetup{
		Stage:         e.Stage,
		Authenticator: info,
		Target:        target,
		Channel:       result.Channel,
		CodeLength:    result.CodeLength,
	}, nil
}

type NodeCreateAuthenticatorOOBSetup struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	Target        string                    `json:"target"`
	Channel       string                    `json:"channel"`
	CodeLength    int                       `json:"code_length"`
}

// GetOOBOTPTarget implements OOBOTPNode.
func (n *NodeCreateAuthenticatorOOBSetup) GetOOBOTPTarget() string {
	return n.Target
}

// GetOOBOTPChannel implements OOBOTPNode.
func (n *NodeCreateAuthenticatorOOBSetup) GetOOBOTPChannel() string {
	return n.Channel
}

// GetOOBOTPOOBType implements OOBOTPNode.
func (n *NodeCreateAuthenticatorOOBSetup) GetOOBOTPOOBType() interaction.OOBType {
	switch n.Stage {
	case authn.AuthenticationStagePrimary:
		return interaction.OOBTypeSetupPrimary
	case authn.AuthenticationStageSecondary:
		return interaction.OOBTypeSetupSecondary
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}
}

// GetOOBOTPCodeLength implements OOBOTPNode.
func (n *NodeCreateAuthenticatorOOBSetup) GetOOBOTPCodeLength() int {
	return n.CodeLength
}

// GetCreateAuthenticatorStage implements CreateAuthenticatorPhoneOTPNode
func (n *NodeCreateAuthenticatorOOBSetup) GetCreateAuthenticatorStage() authn.AuthenticationStage {
	return n.Stage
}

// GetSelectedPhoneNumberForPhoneOTP implements CreateAuthenticatorPhoneOTPNode
func (n *NodeCreateAuthenticatorOOBSetup) GetSelectedPhoneNumberForPhoneOTP() string {
	if n.Channel == string(model.AuthenticatorOOBChannelSMS) {
		return n.Target
	}
	return ""
}

func (n *NodeCreateAuthenticatorOOBSetup) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorOOBSetup) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorOOBSetup) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeOOBResendCode{
			Stage:            n.Stage,
			IsAuthenticating: false,
			Authenticator:    n.Authenticator,
			OTPForm:          otp.FormCode,
		},
		&EdgeCreateAuthenticatorOOB{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
