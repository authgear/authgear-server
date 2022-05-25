package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
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
		CodeID:          verification.NewCodeID(),
		RequestedByUser: e.RequestedByUser,
	}
	result, err := node.SendCode(ctx)
	if err != nil {
		return nil, err
	}

	node.Channel = result.Channel
	node.Target = result.Target
	node.CodeLength = result.CodeLength
	return node, nil
}

type EdgeVerifyIdentityResume struct {
	Code     *verification.Code
	Identity *identity.Info
}

func (e *EdgeVerifyIdentityResume) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	r := e.Code.SendResult()
	return &NodeVerifyIdentity{
		Identity:   e.Identity,
		CodeID:     e.Code.ID,
		Target:     r.Target,
		Channel:    r.Channel,
		CodeLength: r.CodeLength,
		// VerifyIdentityResume is always requested by user.
		RequestedByUser: true,
	}, nil
}

type NodeVerifyIdentity struct {
	Identity        *identity.Info `json:"identity"`
	CodeID          string         `json:"code_id"`
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
		&EdgeVerifyIdentityCheckCode{Identity: n.Identity, ID: n.CodeID},
		&EdgeVerifyIdentityResendCode{Node: n},
	}, nil
}

func (n *NodeVerifyIdentity) SendCode(ctx *interaction.Context) (*otp.CodeSendResult, error) {
	code, err := ctx.Verification.GetCode(n.CodeID)
	if errors.Is(err, verification.ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code == nil || ctx.Clock.NowUTC().After(code.ExpireAt) {
		code, err = ctx.Verification.CreateNewCode(
			n.CodeID,
			n.Identity,
			ctx.WebSessionID,
			n.RequestedByUser,
		)
		if err != nil {
			return nil, err
		}
	}

	// disallow sending sms verification code if phone identity is disabled
	fc := ctx.FeatureConfig
	if config.LoginIDKeyType(code.LoginIDType) == config.LoginIDKeyTypePhone {
		if fc.Identity.LoginID.Types.Phone.Disabled {
			return nil, feature.ErrFeatureDisabledSendingSMS
		}
	}

	err = ctx.RateLimiter.TakeToken(interaction.SendVerificationCodeRateLimitBucket(code.LoginID))
	if err != nil {
		return nil, err
	}

	err = ctx.VerificationCodeSender.SendCode(code)
	if err != nil {
		return nil, err
	}

	return code.SendResult(), nil
}

type InputVerifyIdentityCheckCode interface {
	GetVerificationCode() string
}

type EdgeVerifyIdentityCheckCode struct {
	Identity *identity.Info
	ID       string
}

func (e *EdgeVerifyIdentityCheckCode) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	var input InputVerifyIdentityCheckCode
	if !interaction.Input(rawInput, &input) {
		return nil, interaction.ErrIncompatibleInput
	}

	code, err := ctx.Verification.VerifyCode(e.ID, input.GetVerificationCode())
	if err != nil {
		return nil, err
	}

	claimName := ""
	switch config.LoginIDKeyType(code.LoginIDType) {
	case config.LoginIDKeyTypeEmail:
		claimName = identity.StandardClaimEmail
	case config.LoginIDKeyTypePhone:
		claimName = identity.StandardClaimPhoneNumber
	case config.LoginIDKeyTypeUsername:
		claimName = identity.StandardClaimPreferredUsername
	default:
		panic("interaction: unexpected login ID key")
	}

	verifiedClaim := ctx.Verification.NewVerifiedClaim(code.UserID, claimName, code.LoginID)
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

	_, err := e.Node.SendCode(ctx)
	if err != nil {
		return nil, err
	}

	return nil, interaction.ErrSameNode
}
