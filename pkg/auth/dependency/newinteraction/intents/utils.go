package intents

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction/nodes"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func needPrimaryAuthn(t authn.IdentityType) bool {
	switch t {
	case authn.IdentityTypeLoginID:
		return true
	case authn.IdentityTypeAnonymous, authn.IdentityTypeOAuth:
		return false
	default:
		panic("interaction: unknown identity type" + t)
	}
}

func firstAuthenticationStage(t authn.IdentityType) newinteraction.AuthenticationStage {
	if needPrimaryAuthn(t) {
		return newinteraction.AuthenticationStagePrimary
	}
	return newinteraction.AuthenticationStageSecondary
}

func mustFindNodeSelectIdentity(graph *newinteraction.Graph) *nodes.NodeSelectIdentityEnd {
	var selectIdentity *nodes.NodeSelectIdentityEnd
	for _, node := range graph.Nodes {
		if node, ok := node.(*nodes.NodeSelectIdentityEnd); ok {
			selectIdentity = node
			break
		}
	}
	if selectIdentity == nil {
		panic("interaction: expect identity already selected")
	}
	return selectIdentity
}
