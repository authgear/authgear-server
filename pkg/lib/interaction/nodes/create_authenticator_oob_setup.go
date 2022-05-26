package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/validation"
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

// nolint: gocyclo
func (e *EdgeCreateAuthenticatorOOBSetup) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var target string
	var channel model.AuthenticatorOOBChannel
	if e.Stage == authn.AuthenticationStagePrimary {
		var input InputCreateAuthenticatorOOBSetupSelect
		matchedInput := interaction.Input(rawInput, &input)
		if !matchedInput && !interaction.IsAdminAPI(rawInput) {
			return nil, interaction.ErrIncompatibleInput
		}
		identityInfo := graph.MustGetUserLastIdentity()
		target = identityInfo.Claims[identity.IdentityClaimLoginIDValue].(string)
		loginIDType := identityInfo.Claims[identity.IdentityClaimLoginIDType].(string)
		switch config.LoginIDKeyType(loginIDType) {
		case config.LoginIDKeyTypePhone:
			channel = model.AuthenticatorOOBChannelSMS
		case config.LoginIDKeyTypeEmail:
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

	// Validate target against channel
	validationCtx := &validation.Context{}
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		err := validation.FormatEmail{AllowName: false}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
		}
	case model.AuthenticatorOOBChannelSMS:
		err := validation.FormatPhone{}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "phone"})
		}
	}

	err := validationCtx.Error("invalid target")
	if err != nil {
		return nil, err
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
			Claims:    map[string]interface{}{},
		}

		// Ignore given OOB target, use channel & target inferred from identity
		loginIDKey := identityInfo.Claims[identity.IdentityClaimLoginIDKey].(string)
		for _, t := range ctx.Config.Identity.LoginID.Keys {
			if t.Key != loginIDKey {
				continue
			}
			switch t.Type {
			case config.LoginIDKeyTypeEmail:
				channel = model.AuthenticatorOOBChannelEmail
				oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
			case config.LoginIDKeyTypePhone:
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
		target = identityInfo.Claims[identity.IdentityClaimLoginIDValue].(string)

	} else {
		userID := graph.MustGetUserID()
		spec = &authenticator.Spec{
			UserID:    userID,
			IsDefault: e.IsDefault,
			Kind:      stageToAuthenticatorKind(e.Stage),
			Claims:    map[string]interface{}{},
		}

		// Normalize the target.
		switch channel {
		case model.AuthenticatorOOBChannelEmail:
			var err error
			target, err = ctx.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypeEmail).Normalize(target)
			if err != nil {
				return nil, err
			}
			oobAuthenticatorType = model.AuthenticatorTypeOOBEmail
		case model.AuthenticatorOOBChannelSMS:
			var err error
			target, err = ctx.LoginIDNormalizerFactory.NormalizerWithLoginIDType(config.LoginIDKeyTypePhone).Normalize(target)
			if err != nil {
				return nil, err
			}
			oobAuthenticatorType = model.AuthenticatorTypeOOBSMS
		default:
			panic("interaction: creating OOB authenticator for invalid channel")
		}
	}

	spec.Type = oobAuthenticatorType
	switch channel {
	case model.AuthenticatorOOBChannelSMS:
		spec.Claims[authenticator.AuthenticatorClaimOOBOTPPhone] = target
	case model.AuthenticatorOOBChannelEmail:
		spec.Claims[authenticator.AuthenticatorClaimOOBOTPEmail] = target
	}

	info, err := ctx.Authenticators.NewWithAuthenticatorID(e.NewAuthenticatorID, spec, "")
	if err != nil {
		return nil, err
	}

	var skipInput interface{ SkipVerification() bool }
	if interaction.Input(rawInput, &skipInput) && skipInput.SkipVerification() {
		// Skip verification of OOB target
		return &NodeCreateAuthenticatorOOB{Stage: e.Stage, Authenticator: info}, nil
	}

	aStatus, err := ctx.Verification.GetAuthenticatorVerificationStatus(info)
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
	}).Do()
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

func (n *NodeCreateAuthenticatorOOBSetup) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorOOBSetup) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorOOBSetup) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeOOBResendCode{
			Stage:            n.Stage,
			IsAuthenticating: false,
			Authenticator:    n.Authenticator,
		},
		&EdgeCreateAuthenticatorOOB{Stage: n.Stage, Authenticator: n.Authenticator},
	}, nil
}
