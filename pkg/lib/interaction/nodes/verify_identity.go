package nodes

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

func init() {
	interaction.RegisterNode(&NodeVerifyIdentity{})
}

type InputVerifyIdentity interface {
	SelectVerifyIdentityViaOOBOTP()
}

type EdgeVerifyIdentity struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (e *EdgeVerifyIdentity) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentity
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	node := &NodeVerifyIdentity{
		Identity:        e.Identity,
		RequestedByUser: e.RequestedByUser,
	}
	result, err := node.SendCode(goCtx, ctx, true)
	if err != nil {
		return nil, err
	}

	node.Channel = result.Channel
	node.Target = result.Target
	node.CodeLength = result.CodeLength
	return node, nil
}

type NodeVerifyIdentity struct {
	Identity        *identity.Info `json:"identity"`
	RequestedByUser bool           `json:"requested_by_user"`

	Channel    string `json:"channel"`
	Target     string `json:"target"`
	CodeLength int    `json:"code_length"`
}

// GetVerificationIdentity implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationIdentity() *identity.Info {
	return n.Identity
}

// GetVerificationCodeChannel implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationCodeChannel() string {
	return n.Channel
}

// GetVerificationCodeTarget implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationCodeTarget() string {
	return n.Target
}

// GetVerificationCodeLength implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationCodeLength() int {
	return n.CodeLength
}

// GetVerificationCodeChannel implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetRequestedByUser() bool {
	return n.RequestedByUser
}

func (n *NodeVerifyIdentity) Prepare(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeVerifyIdentity) GetEffects(goCtx context.Context) ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeVerifyIdentity) DeriveEdges(goCtx context.Context, graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeVerifyIdentityCheckCode{Identity: n.Identity},
		&EdgeVerifyIdentityResendCode{Node: n},
	}, nil
}

func (n *NodeVerifyIdentity) SendCode(goCtx context.Context, ctx *interaction.Context, ignoreRatelimitError bool) (*SendOOBCodeResult, error) {
	loginIDType := n.Identity.LoginID.LoginIDType
	channel, target := n.Identity.LoginID.Deprecated_ToChannelTarget()

	result := &SendOOBCodeResult{
		Channel:    string(channel),
		Target:     target,
		CodeLength: otp.FormCode.CodeLength(),
	}

	// disallow sending sms verification code if phone identity is disabled
	fc := ctx.FeatureConfig
	if model.LoginIDKeyType(loginIDType) == model.LoginIDKeyTypePhone {
		if fc.Identity.LoginID.Types.Phone.Disabled {
			return nil, feature.ErrFeatureDisabledSendingSMS
		}
	}

	code, err := ctx.OTPCodeService.GenerateOTP(goCtx,
		otp.KindVerification(ctx.Config, channel),
		target,
		otp.FormCode,
		&otp.GenerateOptions{WebSessionID: ctx.WebSessionID},
	)
	if ignoreRatelimitError && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	err = ctx.OTPSender.Send(
		goCtx,
		otp.SendOptions{
			Channel: channel,
			Target:  target,
			Form:    otp.FormCode,
			Type:    translation.MessageTypeVerification,
			OTP:     code,
		},
	)
	if ignoreRatelimitError && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	return result, nil
}

type InputVerifyIdentityCheckCode interface {
	GetVerificationCode() string
}

type EdgeVerifyIdentityCheckCode struct {
	Identity *identity.Info
}

func (e *EdgeVerifyIdentityCheckCode) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityCheckCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}
	loginIDModel := e.Identity.LoginID
	channel, target := loginIDModel.Deprecated_ToChannelTarget()

	err := ctx.OTPCodeService.VerifyOTP(goCtx,
		otp.KindVerification(ctx.Config, channel),
		target,
		input.GetVerificationCode(),
		&otp.VerifyOptions{UserID: e.Identity.UserID},
	)
	if apierrors.IsKind(err, otp.InvalidOTPCode) {
		return nil, verification.ErrInvalidVerificationCode
	} else if err != nil {
		return nil, err
	}

	var claimName model.ClaimName
	claimName, ok := model.GetLoginIDKeyTypeClaim(loginIDModel.LoginIDType)
	if !ok {
		panic("interaction: unexpected login ID key")
	}

	verifiedClaim := ctx.Verification.NewVerifiedClaim(goCtx, loginIDModel.UserID, string(claimName), target)
	return &NodeEnsureVerificationEnd{
		Identity:         e.Identity,
		NewVerifiedClaim: verifiedClaim,
	}, nil
}

type InputVerifyIdentityResendCode interface {
	DoResend()
}

type EdgeVerifyIdentityResendCode struct {
	Node *NodeVerifyIdentity
}

func (e *EdgeVerifyIdentityResendCode) Instantiate(goCtx context.Context, ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityResendCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	_, err := e.Node.SendCode(goCtx, ctx, false)
	if err != nil {
		return nil, err
	}

	return nil, interaction.ErrSameNode
}
