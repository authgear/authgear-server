package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/otp"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationOOBTrigger{})
}

type InputAuthenticationOOBTrigger interface {
	GetOOBTarget() string
}

type EdgeAuthenticationOOBTrigger struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationOOBTrigger) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputAuthenticationOOBTrigger)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	identityInfo, infos, err := getAuthenticators(ctx, graph, e.Stage, authn.AuthenticatorTypeOOB)
	if err != nil {
		return nil, err
	}

	var targetInfo *authenticator.Info
	if len(infos) == 1 {
		// Select the only authenticator by default
		targetInfo = infos[0]
	} else if target := input.GetOOBTarget(); target != "" {
		// Match authenticators with OOB target
		for _, info := range infos {
			switch info.Props[authenticator.AuthenticatorPropOOBOTPChannelType] {
			case authn.AuthenticatorOOBChannelEmail:
				if info.Props[authenticator.AuthenticatorPropOOBOTPEmail] == target {
					targetInfo = info
				}
			case authn.AuthenticatorOOBChannelSMS:
				if info.Props[authenticator.AuthenticatorPropOOBOTPPhone] == target {
					targetInfo = info
				}
			}
		}
	}

	if targetInfo == nil {
		return nil, newinteraction.ErrInvalidCredentials
	}

	secret, err := otp.GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}

	result, err := sendOOBCode(ctx, e.Stage, otp.OOBOperationTypeAuthenticate, identityInfo, targetInfo, secret)
	if err != nil {
		return nil, err
	}

	return &NodeAuthenticationOOBTrigger{
		Stage:         e.Stage,
		Identity:      identityInfo,
		Authenticator: targetInfo,
		Secret:        secret,
		Channel:       result.Channel,
		CodeLength:    result.CodeLength,
		SendCooldown:  result.SendCooldown,
	}, nil
}

type NodeAuthenticationOOBTrigger struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Identity      *identity.Info                     `json:"identity"`
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
			Stage:         n.Stage,
			Operation:     otp.OOBOperationTypeAuthenticate,
			Identity:      n.Identity,
			Authenticator: n.Authenticator,
			Secret:        n.Secret,
		},
		&EdgeAuthenticationOOB{Stage: n.Stage, Authenticator: n.Authenticator, Secret: n.Secret},
	}, nil
}
