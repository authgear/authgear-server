package nodes

import (
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/feature/verification"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
)

func init() {
	interaction.RegisterNode(&NodeEnsureVerificationBegin{})
	interaction.RegisterNode(&NodeEnsureVerificationEnd{})
}

type EdgeEnsureVerificationBegin struct {
	Identity        *identity.Info
	RequestedByUser bool
}

func (e *EdgeEnsureVerificationBegin) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	skipVerification := false
	var skipInput interface{ SkipVerification() bool }
	if interaction.Input(rawInput, &skipInput) && skipInput.SkipVerification() {
		skipVerification = true
	}

	return &NodeEnsureVerificationBegin{
		Identity:         e.Identity,
		RequestedByUser:  e.RequestedByUser,
		SkipVerification: skipVerification,
		PhoneOTPMode:     ctx.Config.Authenticator.OOB.SMS.PhoneOTPMode,
	}, nil
}

type NodeEnsureVerificationBegin struct {
	Identity           *identity.Info                   `json:"identity"`
	RequestedByUser    bool                             `json:"requested_by_user"`
	SkipVerification   bool                             `json:"skip_verification"`
	PhoneOTPMode       config.AuthenticatorPhoneOTPMode `json:"phone_otp_mode"`
	VerificationStatus verification.Status              `json:"-"`
}

// GetVerifyIdentityEdges implements EnsureVerificationBeginNode
func (n *NodeEnsureVerificationBegin) GetVerifyIdentityEdges() ([]interaction.Edge, error) {
	return n.deriveEdges()
}

func (n *NodeEnsureVerificationBegin) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	claims, err := ctx.Verification.GetIdentityVerificationStatus(n.Identity)
	if err != nil {
		return err
	}

	// TODO(verification): handle multiple verifiable claims per identity
	if len(claims) > 0 {
		n.VerificationStatus = claims[0].Status
	} else {
		n.VerificationStatus = verification.StatusDisabled
	}
	return nil
}

func (n *NodeEnsureVerificationBegin) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeEnsureVerificationBegin) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return n.deriveEdges()
}

func (n *NodeEnsureVerificationBegin) deriveEdges() ([]interaction.Edge, error) {
	isPhoneIdentity := ensurePhoneLoginIDIdentity(n.Identity) == nil
	verifyIdentityEdges := func() (edges []interaction.Edge) {
		if n.PhoneOTPMode.IsWhatsappEnabled() && isPhoneIdentity {
			edges = append(edges, &EdgeVerifyIdentityViaWhatsapp{
				Identity:        n.Identity,
				RequestedByUser: n.RequestedByUser,
			})
		}

		edges = append(edges, &EdgeVerifyIdentity{
			Identity:        n.Identity,
			RequestedByUser: n.RequestedByUser,
		})
		return edges
	}
	switch n.VerificationStatus {
	case verification.StatusDisabled, verification.StatusVerified:
		break
	case verification.StatusPending:
		if n.RequestedByUser && !n.SkipVerification {
			return verifyIdentityEdges(), nil
		}
	case verification.StatusRequired:
		if !n.SkipVerification {
			return verifyIdentityEdges(), nil
		}
	}

	return []interaction.Edge{&EdgeEnsureVerificationEnd{Identity: n.Identity}}, nil
}

type EdgeEnsureVerificationEnd struct {
	Identity *identity.Info
}

func (e *EdgeEnsureVerificationEnd) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	return &NodeEnsureVerificationEnd{
		Identity: e.Identity,
	}, nil
}

type NodeEnsureVerificationEnd struct {
	Identity         *identity.Info      `json:"identity"`
	NewVerifiedClaim *verification.Claim `json:"new_verified_claim,omitempty"`
}

func (n *NodeEnsureVerificationEnd) Prepare(ctx *interaction.Context, graph *interaction.Graph) error {
	return nil
}

func (n *NodeEnsureVerificationEnd) GetEffects() ([]interaction.Effect, error) {
	return nil, nil
}

func (n *NodeEnsureVerificationEnd) DeriveEdges(graph *interaction.Graph) ([]interaction.Edge, error) {
	return graph.Intent.DeriveEdgesForNode(graph, n)
}
