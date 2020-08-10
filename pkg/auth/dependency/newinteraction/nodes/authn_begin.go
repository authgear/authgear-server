package nodes

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func init() {
	newinteraction.RegisterNode(&NodeAuthenticationBegin{})
}

type EdgeAuthenticationBegin struct {
	Stage newinteraction.AuthenticationStage
}

func (e *EdgeAuthenticationBegin) Instantiate(ctx *newinteraction.Context, graph *newinteraction.Graph, input interface{}) (newinteraction.Node, error) {
	return &NodeAuthenticationBegin{
		Stage: e.Stage,
	}, nil
}

type NodeAuthenticationBegin struct {
	Stage                newinteraction.AuthenticationStage `json:"stage"`
	AuthenticationConfig *config.AuthenticationConfig       `json:"-"`
	Authenticators       []*authenticator.Info              `json:"-"`
}

func (n *NodeAuthenticationBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.AuthenticationConfig = ctx.Config.Authentication
	n.Authenticators = ais
	return nil
}

func (n *NodeAuthenticationBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	var err error
	var availableAuthenticators []*authenticator.Info
	identityInfo := graph.MustGetUserLastIdentity()

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
			authenticator.KeepPrimaryAuthenticatorOfIdentity(identityInfo),
		)
		if err != nil {
			return nil, err
		}
		availableAuthenticators = newinteraction.SortAuthenticators(availableAuthenticators, n.AuthenticationConfig.PrimaryAuthenticators)
	case newinteraction.AuthenticationStageSecondary:
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
		)
		if err != nil {
			return nil, err
		}
	default:
		panic("interaction: unknown authentication stage: " + n.Stage)
	}

	passwords := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypePassword),
	)
	totps := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeTOTP),
	)
	oobs := filterAuthenticators(
		availableAuthenticators,
		authenticator.KeepType(authn.AuthenticatorTypeOOB),
	)

	if len(passwords) > 0 {
		edges = append(edges, &EdgeAuthenticationPassword{
			Stage:          n.Stage,
			Authenticators: passwords,
		})
	}

	if len(totps) > 0 {
		edges = append(edges, &EdgeAuthenticationTOTP{
			Stage:          n.Stage,
			Authenticators: totps,
		})
	}

	if len(oobs) > 0 {
		edges = append(edges, &EdgeAuthenticationOOBTrigger{
			Stage:          n.Stage,
			Authenticators: oobs,
		})
	}

	// No authenticators found, skip the authentication stage
	if len(edges) == 0 {
		edges = append(edges, &EdgeAuthenticationEnd{
			Stage:    n.Stage,
			Optional: true,
		})
	}

	// TODO(interaction): support choosing authenticator to use
	return edges[:1], nil
}
