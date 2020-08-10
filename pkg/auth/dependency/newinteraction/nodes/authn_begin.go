package nodes

import (
	"fmt"

	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/auth/dependency/identity"
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
	Identity             *identity.Info                     `json:"-"`
	AuthenticationConfig *config.AuthenticationConfig       `json:"-"`
	Authenticators       []*authenticator.Info              `json:"-"`
}

func (n *NodeAuthenticationBegin) Prepare(ctx *newinteraction.Context, graph *newinteraction.Graph) error {
	ais, err := ctx.Authenticators.List(graph.MustGetUserID())
	if err != nil {
		return err
	}

	n.Identity = graph.MustGetUserLastIdentity()
	n.AuthenticationConfig = ctx.Config.Authentication
	n.Authenticators = ais
	return nil
}

func (n *NodeAuthenticationBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationBegin) DeriveEdges(graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	return n.GetAuthenticationEdges(), nil
}

// GetAuthenticationEdges implements AuthenticationBeginNode.
func (n *NodeAuthenticationBegin) GetAuthenticationEdges() []newinteraction.Edge {
	var edges []newinteraction.Edge
	var availableAuthenticators []*authenticator.Info
	var preferred []authn.AuthenticatorType

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		preferred = n.AuthenticationConfig.PrimaryAuthenticators
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
			authenticator.KeepPrimaryAuthenticatorOfIdentity(n.Identity),
		)
	case newinteraction.AuthenticationStageSecondary:
		preferred = n.AuthenticationConfig.SecondaryAuthenticators
		availableAuthenticators = filterAuthenticators(
			n.Authenticators,
			authenticator.KeepTag(authenticator.TagSecondaryAuthenticator),
		)
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
		return edges
	}

	newinteraction.SortAuthenticators(
		preferred,
		edges,
		func(i int) authn.AuthenticatorType {
			edge := edges[i]
			switch edge.(type) {
			case *EdgeAuthenticationPassword:
				return authn.AuthenticatorTypePassword
			case *EdgeAuthenticationTOTP:
				return authn.AuthenticatorTypeTOTP
			case *EdgeAuthenticationOOBTrigger:
				return authn.AuthenticatorTypeOOB
			default:
				panic(fmt.Sprintf("interaction: unknown edge: %T", edge))
			}
		},
	)

	return edges
}
