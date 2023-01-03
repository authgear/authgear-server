package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/httputil"
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

func (n *NodeVerifyIdentity) SendCode(ctx *interaction.Context, ignoreRatelimitError bool) (*otp.CodeSendResult, error) {
	loginIDType := n.Identity.LoginID.LoginIDType
	loginID := n.Identity.LoginID.LoginID

	var channel model.AuthenticatorOOBChannel
	var target string
	switch loginIDType {
	case model.LoginIDKeyTypePhone:
		channel = model.AuthenticatorOOBChannelSMS
		target = loginID
	case model.LoginIDKeyTypeEmail:
		channel = model.AuthenticatorOOBChannelEmail
		target = loginID
	default:
		panic("node: incompatible authenticator type for sending oob code: " + loginIDType)
	}

	code, err := ctx.OTPCodeService.GenerateCode(
		target,
		ctx.Clock.NowUTC().Add(ctx.Config.Verification.CodeExpiry.Duration()))
	if err != nil {
		return nil, err
	}

	// disallow sending sms verification code if phone identity is disabled
	fc := ctx.FeatureConfig
	if model.LoginIDKeyType(loginIDType) == model.LoginIDKeyTypePhone {
		if fc.Identity.LoginID.Types.Phone.Disabled {
			return nil, feature.ErrFeatureDisabledSendingSMS
		}
	}

	result := &otp.CodeSendResult{
		Channel:    string(channel),
		Target:     target,
		CodeLength: len(code.Code),
	}
	err = ctx.RateLimiter.TakeToken(interaction.AntiSpamSendVerificationCodeBucket(loginID))
	if ignoreRatelimitError && errors.Is(err, ratelimit.ErrTooManyRequests) {
		// Ignore the rate limit error and do NOT send the code.
		return result, nil
	} else if err != nil {
		return nil, err
	}

	err = ctx.OOBCodeSender.SendCode(channel, target, code.Code, otp.MessageTypeVerification)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type InputVerifyIdentityCheckCode interface {
	GetVerificationCode() string
}

type RateLimiter interface {
	TakeToken(bucket ratelimit.Bucket) error
}

type EdgeVerifyIdentityCheckCode struct {
	RemoteIP    httputil.RemoteIP
	Identity    *identity.Info
	RateLimiter RateLimiter
}

func (e *EdgeVerifyIdentityCheckCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityCheckCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}
	loginIDModel := e.Identity.LoginID

	var target string
	switch loginIDModel.LoginIDType {
	case model.LoginIDKeyTypePhone, model.LoginIDKeyTypeEmail:
		target = loginIDModel.LoginID
	default:
		panic("interaction: incompatible logic ID type for oob code: " + loginIDModel.LoginIDType)
	}

	err := e.RateLimiter.TakeToken(verification.AutiBruteForceVerifyBucket(string(e.RemoteIP)))
	if err != nil {
		return nil, err
	}
	err = ctx.OTPCodeService.VerifyCode(target, input.GetVerificationCode())
	if err != nil {
		return nil, err
	}

	var claimName model.ClaimName
	switch model.LoginIDKeyType(loginIDModel.LoginIDType) {
	case model.LoginIDKeyTypeEmail:
		claimName = model.ClaimEmail
	case model.LoginIDKeyTypePhone:
		claimName = model.ClaimPhoneNumber
	case model.LoginIDKeyTypeUsername:
		claimName = model.ClaimPreferredUsername
	default:
		panic("interaction: unexpected login ID key")
	}

	verifiedClaim := ctx.Verification.NewVerifiedClaim(loginIDModel.UserID, string(claimName), loginIDModel.LoginID)
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
