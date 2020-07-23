package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity/loginid"
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

	// When targetInfo is nil: continue and act as if there is one,
	// prevent enumeration of user's authenticators

	secret, err := otp.GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}

	node := &NodeAuthenticationOOBTrigger{
		Stage:         e.Stage,
		Identity:      identityInfo,
		Authenticator: targetInfo,
		Secret:        secret,
	}
	err = node.sendOOBCode(ctx, graph)
	if err != nil {
		return nil, err
	}

	return node, nil
}

type NodeAuthenticationOOBTrigger struct {
	Stage         newinteraction.AuthenticationStage `json:"stage"`
	Identity      *identity.Info                     `json:"identity"`
	Authenticator *authenticator.Info                `json:"authenticator"`
	Secret        string                             `json:"secret"`
}

func (n *NodeAuthenticationOOBTrigger) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationOOBTrigger) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return []newinteraction.Edge{
		&EdgeAuthenticationOOBResend{Node: n},
		&EdgeAuthenticationOOB{Stage: n.Stage, Authenticator: n.Authenticator, Secret: n.Secret},
	}, nil
}

func (n *NodeAuthenticationOOBTrigger) sendOOBCode(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	if n.Authenticator == nil {
		return nil
	}

	// TODO(interaction): handle rate limits

	channel := authn.AuthenticatorOOBChannel(n.Authenticator.Props[authenticator.AuthenticatorPropOOBOTPChannelType].(string))

	var operation otp.OOBOperationType
	var loginID *loginid.LoginID
	if n.Stage == newinteraction.AuthenticationStagePrimary {
		// Primary OOB authenticators is bound to login ID identities:
		// Extract login ID from the bound identity.
		operation = otp.OOBOperationTypePrimaryAuth
		if n.Identity != nil {
			loginID = &loginid.LoginID{
				Key:   n.Identity.Claims[identity.IdentityClaimLoginIDKey].(string),
				Value: n.Identity.Claims[identity.IdentityClaimLoginIDValue].(string),
			}
		}
	} else {
		// Secondary OOB authenticators is not bound to login ID identities.
		operation = otp.OOBOperationTypeSecondaryAuth
		loginID = nil
	}

	// Use a placeholder login ID if no bound login ID identity
	if loginID == nil {
		loginID = &loginid.LoginID{}
		switch channel {
		case authn.AuthenticatorOOBChannelSMS:
			loginID.Value = n.Authenticator.Props[authenticator.AuthenticatorPropOOBOTPPhone].(string)
		case authn.AuthenticatorOOBChannelEmail:
			loginID.Value = n.Authenticator.Props[authenticator.AuthenticatorPropOOBOTPEmail].(string)
		}
	}

	// TODO(interaction): determine intent type
	var origin = otp.MessageOriginLogin

	code := ctx.OOBAuthenticators.GenerateCode(n.Secret, channel)
	return ctx.OOBAuthenticators.SendCode(channel, loginID, code, origin, operation)
}

type InputAuthenticationOOBResend interface {
	DoResend() bool
}

type EdgeAuthenticationOOBResend struct {
	Node *NodeAuthenticationOOBTrigger
}

func (e *EdgeAuthenticationOOBResend) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, rawInput interface{}) (newinteraction.Node, error) {
	_, ok := rawInput.(InputAuthenticationOOBResend)
	if !ok {
		return nil, newinteraction.ErrIncompatibleInput
	}

	err := e.Node.sendOOBCode(ctx, graph)
	if err != nil {
		return nil, err
	}

	return nil, newinteraction.ErrSameNode
}
