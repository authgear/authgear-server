package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
	"github.com/authgear/authgear-server/pkg/otp"
)

func init() {
	newinteraction.RegisterNode(&NodeCreateAuthenticatorOOBSetup{})
}

type InputCreateAuthenticatorOOBSetup interface {
	GetOOBChannel() authn.AuthenticatorOOBChannel
	GetOOBTarget() string
}

type EdgeCreateAuthenticatorOOBSetup struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeCreateAuthenticatorOOBSetup) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	input, ok := rawInput.(InputCreateAuthenticatorOOBSetup)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}
	channel := input.GetOOBChannel()
	if channel == "" {
		return nil, newinteraction.ErrIncompatibleInput
	}

	var spec *authenticator.Spec
	var identityInfo *identity.Info
	target := input.GetOOBTarget()
	if e.Stage == newinteraction.AuthenticationStagePrimary {
		// Primary OOB authenticators must be bound to login ID identity
		identityInfo = graph.MustGetUserLastIdentity()
		if identityInfo.Type != authn.IdentityTypeLoginID {
			panic("interaction: OOB authenticator identity must be login ID")
		}

		spec = &authenticator.Spec{
			UserID: identityInfo.UserID,
			Type:   authn.AuthenticatorTypeOOB,
			Props:  map[string]interface{}{},
		}

		// Ignore given OOB target, use channel & target inferred from identity
		loginIDKey := identityInfo.Claims[identity.IdentityClaimLoginIDKey].(string)
		for _, t := range ctx.Config.Identity.LoginID.Keys {
			if t.Key != loginIDKey {
				continue
			}
			switch t.Type {
			case config.LoginIDKeyTypeEmail:
				channel = authn.AuthenticatorOOBChannelEmail
			case config.LoginIDKeyTypePhone:
				channel = authn.AuthenticatorOOBChannelSMS
			default:
				return nil, newinteraction.ConfigurationViolated.New("this login ID cannot be used for OOB authentication")
			}
			break
		}
		target = identityInfo.Claims[identity.IdentityClaimLoginIDValue].(string)

	} else {
		userID := graph.MustGetUserID()
		spec = &authenticator.Spec{
			UserID: userID,
			Type:   authn.AuthenticatorTypeOOB,
			Props:  map[string]interface{}{},
		}
	}

	spec.Props[authenticator.AuthenticatorPropOOBOTPChannelType] = string(channel)
	switch channel {
	case authn.AuthenticatorOOBChannelSMS:
		spec.Props[authenticator.AuthenticatorPropOOBOTPPhone] = target
	case authn.AuthenticatorOOBChannelEmail:
		spec.Props[authenticator.AuthenticatorPropOOBOTPEmail] = target
	}

	infos, err := ctx.Authenticators.New(spec, "")
	if err != nil {
		return nil, err
	}

	if len(infos) != 1 {
		panic("interaction: unexpected number of new OOB authenticators")
	}

	secret, err := otp.GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}

	result, err := sendOOBCode(ctx, e.Stage, otp.OOBOperationTypeSetup, identityInfo, infos[0], secret)
	if err != nil {
		return nil, err
	}

	return &NodeCreateAuthenticatorOOBSetup{
		Stage:         e.Stage,
		Identity:      identityInfo,
		Authenticator: infos[0],
		Secret:        secret,
		Channel:       result.Channel,
		CodeLength:    result.CodeLength,
		SendCooldown:  result.SendCooldown,
	}, nil
}

type NodeCreateAuthenticatorOOBSetup struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Identity      *identity.Info                     `json:"identity"`
	Authenticator *authenticator.Info                `json:"authenticator"`
	Secret        string                             `json:"secret"`
	Channel       string                             `json:"channel"`
	CodeLength    int                                `json:"code_length"`
	SendCooldown  int                                `json:"send_cooldown"`
}

// GetOOBOTPChannel implements OOBOTPNode.
func (n *NodeCreateAuthenticatorOOBSetup) GetOOBOTPChannel() string {
	return n.Channel
}

// GetOOBOTPCodeSendCooldown implements OOBOTPNode.
func (n *NodeCreateAuthenticatorOOBSetup) GetOOBOTPCodeSendCooldown() int {
	return n.SendCooldown
}

// GetOOBOTPCodeLength implements OOBOTPNode.
func (n *NodeCreateAuthenticatorOOBSetup) GetOOBOTPCodeLength() int {
	return n.CodeLength
}

func (n *NodeCreateAuthenticatorOOBSetup) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeCreateAuthenticatorOOBSetup) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeOOBResendCode{
			Stage:         n.Stage,
			Operation:     otp.OOBOperationTypeSetup,
			Identity:      n.Identity,
			Authenticator: n.Authenticator,
			Secret:        n.Secret,
		},
		&EdgeCreateAuthenticatorOOB{Stage: n.Stage, Authenticator: n.Authenticator, Secret: n.Secret},
	}, nil
}
