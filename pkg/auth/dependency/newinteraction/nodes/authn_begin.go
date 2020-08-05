package nodes

import (
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
	Stage newinteraction.AuthenticationStage `json:"stage"`
}

func (n *NodeAuthenticationBegin) Apply(perform func(eff newinteraction.Effect) error, graph *newinteraction.Graph) error {
	return nil
}

func (n *NodeAuthenticationBegin) DeriveEdges(ctx *newinteraction.Context, graph *newinteraction.Graph) ([]newinteraction.Edge, error) {
	var edges []newinteraction.Edge
	var err error
	var availableAuthenticators []*authenticator.Info
	identityInfo := graph.MustGetUserLastIdentity()

	switch n.Stage {
	case newinteraction.AuthenticationStagePrimary:
		availableAuthenticators, err = ctx.Authenticators.List(
			identityInfo.UserID,
			authenticator.KeepTag(authenticator.TagPrimaryAuthenticator),
			authenticator.KeepPrimaryAuthenticatorOfIdentity(identityInfo),
		)
		if err != nil {
			return nil, err
		}
		availableAuthenticators = newinteraction.SortAuthenticators(availableAuthenticators, ctx.Config.Authentication.PrimaryAuthenticators)
	case newinteraction.AuthenticationStageSecondary:
		availableAuthenticators, err = ctx.Authenticators.List(
			identityInfo.UserID,
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
		// FIXME(mfa): Change AuthenticatorService API to make its Authenticate taking infos.
		edges = append(edges, &EdgeAuthenticationPassword{Stage: n.Stage})
	}

	if len(totps) > 0 {
		// FIXME(mfa): Change AuthenticatorService API to make its Authenticate taking infos.
		edges = append(edges, &EdgeAuthenticationTOTP{Stage: n.Stage})
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
