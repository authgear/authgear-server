package nodes

import (
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

func (e *EdgeVerifyIdentity) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentity
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	node := &NodeVerifyIdentity{
		Identity:        e.Identity,
		RequestedByUser: e.RequestedByUser,
	}
	result, err := node.SendCode(ctx, true)
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

func (n *NodeVerifyIdentity) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeVerifyIdentity) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeVerifyIdentity) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return []interaction.Edge{
		&EdgeVerifyIdentityCheckCode{Identity: n.Identity},
		&EdgeVerifyIdentityResendCode{Node: n},
	}, nil
}

func (n *NodeVerifyIdentity) SendCode(ctx *interaction.Context, ignoreRatelimitError bool) (*SendOOBCodeResult, error) {
	loginIDType := n.Identity.LoginID.LoginIDType
	channel, target := n.Identity.LoginID.ToChannelTarget()

	result := &SendOOBCodeResult{
		Channel:    string(channel),
		Target:     target,
		CodeLength: otp.FormCode.CodeLength(),
	}

	msg, err := ctx.OTPSender.Prepare(channel, target, otp.FormCode, translation.MessageTypeVerification)
	if ignoreRatelimitError && apierrors.IsKind(err, ratelimit.RateLimited) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}
	defer msg.Close()

	code, err := ctx.OTPCodeService.GenerateOTP(
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

	// disallow sending sms verification code if phone identity is disabled
	fc := ctx.FeatureConfig
	if model.LoginIDKeyType(loginIDType) == model.LoginIDKeyTypePhone {
		if fc.Identity.LoginID.Types.Phone.Disabled {
			return nil, feature.ErrFeatureDisabledSendingSMS
		}
	}

	err = ctx.OTPSender.Send(msg, otp.SendOptions{OTP: code})
	if err != nil {
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

func (e *EdgeVerifyIdentityCheckCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityCheckCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}
	loginIDModel := e.Identity.LoginID
	channel, target := loginIDModel.ToChannelTarget()

	err := ctx.OTPCodeService.VerifyOTP(
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

	verifiedClaim := ctx.Verification.NewVerifiedClaim(loginIDModel.UserID, string(claimName), target)
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

func (e *EdgeVerifyIdentityResendCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityResendCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	_, err := e.Node.SendCode(ctx, false)
	if err != nil {
		return nil, err
	}

	return nil, interaction.ErrSameNode
}
