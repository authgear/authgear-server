package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeCreateAuthenticatorMagicLinkOTPSetup{})
}

type InputCreateAuthenticatorMagicLinkOTPSetup interface {
	GetMagicLinkTarget() string
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
		panic("Magic link as primary authenticator is not yet support")
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
			Email: input.GetMagicLinkTarget(),
		},
		MagicLinkOTP: &authenticator.MagicLinkOTPSpec{
			Email: input.GetMagicLinkTarget(),
		},
	}

	info, err := ctx.Authenticators.NewWithAuthenticatorID(e.NewAuthenticatorID, spec)
	if err != nil {
		return nil, err
	}

	var skipInput interface{ SkipVerification() bool }
	if interaction.Input(rawInput, &skipInput) && skipInput.SkipVerification() {
		// Admin skip verify MagicLink otp and create OOB authenticator directly
		return &NodeCreateAuthenticatorOOB{Stage: e.Stage, Authenticator: info}, nil
	}

	// TODO(newman): Create and send token

	return &NodeCreateAuthenticatorMagicLinkOTPSetup{
		Stage:         e.Stage,
		Authenticator: info,
		MagicLinkOTP:  "fixme",
		Channel:       "email",
		Target:        input.GetMagicLinkTarget(),
	}, nil
}

type NodeCreateAuthenticatorMagicLinkOTPSetup struct {
	Stage         authn.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info       `json:"authenticator"`
	MagicLinkOTP  string                    `json:"magic_link_otp"`
	Target        string                    `json:"target"`
	Channel       string                    `json:"channel"`
}

// GetMagicLinkOTPTarget implements MagicLinkOTPNode.
func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetMagicLinkOTPTarget() string {
	return n.Target
}

// GetPhone implements MagicLinkOTPNode.
func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetMagicLinkOTPChannel() string {
	return n.Channel
}

// GetCreateAuthenticatorStage implements CreateAuthenticatorPhoneOTPNode
func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetCreateAuthenticatorStage() authn.AuthenticationStage {
	return n.Stage
}

func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeCreateAuthenticatorMagicLinkOTPSetup) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	edges := []interaction.Edge{
		&EdgeCreateAuthenticatorMagicLinkOTP{Stage: n.Stage, Authenticator: n.Authenticator},
	}
	return edges, nil
}
