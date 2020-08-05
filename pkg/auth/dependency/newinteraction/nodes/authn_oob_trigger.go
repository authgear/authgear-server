package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/otp"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationOOBTrigger{})
}

type EdgeAuthenticationOOBTrigger struct {
	Stage          newinteraction.AuthenticationStage
	Authenticators []*authenticator.Info
}

func (e *EdgeAuthenticationOOBTrigger) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	// FIXME(mfa): Support switching another authenticator.
	targetInfo := e.Authenticators[0]

	secret, err := otp.GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}

	result, err := sendOOBCode(ctx, e.Stage, true, targetInfo, secret)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationOOBTrigger{
		Stage:         e.Stage,
		Authenticator: targetInfo,
		Secret:        secret,
		Channel:       result.Channel,
		CodeLength:    result.CodeLength,
		SendCooldown:  result.SendCooldown,
	}, nil
}

type NodeAuthenticationOOBTrigger struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Authenticator *authenticator.Info                `json:"authenticator"`
	Secret        string                             `json:"secret"`
	Channel       string                             `json:"channel"`
	CodeLength    int                                `json:"code_length"`
	SendCooldown  int                                `json:"send_cooldown"`
}

// GetOOBOTPChannel implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPChannel() string {
	return n.Channel
}

// GetOOBOTPCodeSendCooldown implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPCodeSendCooldown() int {
	return n.SendCooldown
}

// GetOOBOTPCodeLength implements OOBOTPNode.
func (n *NodeAuthenticationOOBTrigger) GetOOBOTPCodeLength() int {
	return n.CodeLength
}

func (n *NodeAuthenticationOOBTrigger) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOBTrigger) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeOOBResendCode{
			Stage:            n.Stage,
			IsAuthenticating: true,
			Authenticator:    n.Authenticator,
			Secret:           n.Secret,
		},
		&EdgeAuthenticationOOB{Stage: n.Stage, Authenticator: n.Authenticator, Secret: n.Secret},
	}, nil
}
