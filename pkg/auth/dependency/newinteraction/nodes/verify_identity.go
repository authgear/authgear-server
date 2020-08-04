package nodes

import (
	"errors"

	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/verification"
	"github.com/authgear/authgear-server/pkg/core/uuid"
	"github.com/authgear/authgear-server/pkg/otp"
)

func init() {
	newinteraction.RegisterNode(&NodeVerifyIdentity{})
}

type EdgeVerifyIdentity struct {
	Identity *identity.Info
}

func (e *EdgeVerifyIdentity) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	node := &NodeVerifyIdentity{
		Identity: e.Identity,
		ID:       uuid.New(),
	}
	result, err := node.SendCode(ctx)
	if err != nil {
		return nil, err
	}

	node.Channel = result.Channel
	node.CodeLength = result.CodeLength
	node.SendCooldown = result.SendCooldown
	return node, nil
}

type NodeVerifyIdentity struct {
	Identity     *identity.Info `json:"identity"`
	ID           string         `json:"id"`
	Channel      string         `json:"channel"`
	CodeLength   int            `json:"code_length"`
	SendCooldown int            `json:"send_cooldown"`
}

// GetVerificationIdentity implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationIdentity() *identity.Info {
	return n.Identity
}

// GetVerificationCodeChannel implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationCodeChannel() string {
	return n.Channel
}

// GetVerificationCodeSendCooldown implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationCodeSendCooldown() int {
	return n.SendCooldown
}

// GetVerificationCodeLength implements VerifyIdentityNode.
func (n *NodeVerifyIdentity) GetVerificationCodeLength() int {
	return n.CodeLength
}

func (n *NodeVerifyIdentity) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeVerifyIdentity) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeVerifyIdentityCheckCode{Identity: n.Identity, ID: n.ID},
		&EdgeVerifyIdentityResendCode{Node: n},
	}, nil
}

func (n *NodeVerifyIdentity) SendCode(ctx *newinteraction.Context) (*otp.CodeSendResult, error) {
	code, err := ctx.Verification.GetCode(n.ID)
	if errors.Is(err, verification.ErrCodeNotFound) {
		code = nil
	} else if err != nil {
		return nil, err
	}

	if code == nil || ctx.Clock.NowUTC().After(code.ExpireAt) {
		code, err = ctx.Verification.CreateNewCode(n.ID, n.Identity)
		if err != nil {
			return nil, err
		}
	}

	// TODO: generate verification link
	result, err := ctx.Verification.SendCode(code, "")
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
	ID       string
}

func (e *EdgeVerifyIdentityCheckCode) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputVerifyIdentityCheckCode)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	code, err := ctx.Verification.VerifyCode(e.ID, input.GetVerificationCode())
	if err != nil {
		return nil, err
	}

	newAuthenticator, err := ctx.Verification.NewVerificationAuthenticator(code)
	if err != nil {
		return nil, err
	}

	return &NodeEnsureVerificationEnd{
		Identity:         e.Identity,
		NewAuthenticator: newAuthenticator,
	}, nil
}

type InputVerifyIdentityResendCode interface {
	DoResend()
}

type EdgeVerifyIdentityResendCode struct {
	Node *NodeVerifyIdentity
}

func (e *EdgeVerifyIdentityResendCode) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	_, ok := rawInput.(InputVerifyIdentityResendCode)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	_, err := e.Node.SendCode(ctx)
	if err != nil {
		return nil, err
	}

	return nil, newinteraction.ErrSameNode
}
