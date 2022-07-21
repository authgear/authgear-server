package nodes

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func getIdentityConflictNode(graph *interaction.Graph) (*NodeCheckIdentityConflict, bool) {
	for _, node := range graph.Nodes {
		if node, ok := node.(*NodeCheckIdentityConflict); ok {
			return node, true
		}
	}
	return nil, false
}

// EdgeTerminal is used to indicate a terminal state of interaction; the
// interaction cannot further, and must be rewound to a previous step to
// continue.
type EdgeTerminal struct{}

func (e *EdgeTerminal) Instantiate(ctx *interaction.Context, graph *interaction.Graph, rawInput interface{}) (interaction.Node, error) {
	// Use ErrIncompatibleInput to 'stuck' the interaction at the current node.
	return nil, interaction.ErrIncompatibleInput
}

type InputAuthenticationStage interface {
	GetAuthenticationStage() authn.AuthenticationStage
}

func identityFillDetails(err error, spec *identity.Spec, otherSpec *identity.Spec) error {
	details := errorutil.Details{}

	if spec != nil {
		details["IdentityTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.Type)
		switch spec.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.Claims[identity.IdentityClaimLoginIDType])
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.Claims[identity.IdentityClaimOAuthProviderType])
		}
	}

	if otherSpec != nil {
		details["IdentityTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.Type)
		switch otherSpec.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.Claims[identity.IdentityClaimLoginIDType])
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.Claims[identity.IdentityClaimOAuthProviderType])
		}
	}

	return errorutil.WithDetails(err, details)
}

func forgotpasswordFillDetails(err error) error {
	details := errorutil.Details{}
	details["IdentityTypeIncoming"] = apierrors.APIErrorDetail.Value(model.IdentityTypeLoginID)
	details["LoginIDTypeIncoming"] = ""
	return errorutil.WithDetails(err, details)
}
