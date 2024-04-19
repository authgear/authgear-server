package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentAccountLinkingOAuth{})
}

type IntentAccountLinkingOAuth struct {
	LinkToUserID        string         `json:"link_to_user_id,omitempty"`
	OAuthIdentitySpec   *identity.Spec `json:"oauth_identity_spec,omitempty"`
	ConflictingIdentity *identity.Info `json:"conflicting_identity,omitempty"` // TODO(tung): Change this to a list
}

var _ authflow.Intent = &IntentAccountLinkingOAuth{}

func (*IntentAccountLinkingOAuth) Kind() string {
	return "IntentAccountLinkingOAuth"
}

func (*IntentAccountLinkingOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0: // Enter the login flow
		return nil, nil
	case 1: // Add NodeDoCreateIdentity
		return nil, nil
	}

	return nil, authflow.ErrEOF
}

func (i *IntentAccountLinkingOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	flowUserID, err := getUserID(flows)
	conflictedUserID := i.ConflictingIdentity.UserID
	if err != nil {
		return nil, err
	}
	if flowUserID != conflictedUserID {
		err = authflow.TraverseFlow(authflow.Traverser{
			NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
				milestone, ok := nodeSimple.(MilestoneSwitchToExistingUser)
				if ok {
					err = milestone.MilestoneSwitchToExistingUser(deps, w, conflictedUserID)
					if err != nil {
						return err
					}
				}
				return nil
			},
			Intent: func(intent authflow.Intent, w *authflow.Flow) error {
				milestone, ok := intent.(MilestoneSwitchToExistingUser)
				if ok {
					err = milestone.MilestoneSwitchToExistingUser(deps, w, conflictedUserID)
					if err != nil {
						return err
					}
				}
				return nil
			},
		}, flows.Root)
		if err != nil {
			return nil, err
		}
		// Use synthetic input to auto select the conflicted identity in the login flow
		return nil, &authflow.ErrorRewriteFlow{
			Intent:         flows.Root.Intent,
			Nodes:          flows.Root.Nodes,
			SyntheticInput: i.createSyntheticInputOAuthConflict(i.OAuthIdentitySpec, i.ConflictingIdentity),
		}
	}

	switch len(flows.Nearest.Nodes) {
	case 0:
		// TODO(tung): Check the config to decide what to do
		flowReference := authflow.FlowReference{
			Type: authflow.FlowTypeLogin,
			// FIXME(tung): This should be read from config
			Name: "default",
		}
		loginIntent := IntentLoginFlow{
			TargetUserID:  conflictedUserID,
			FlowReference: flowReference,
		}
		return authflow.NewSubFlow(&loginIntent), nil
	case 1:
		info, _, err := newIdentityInfo(deps, i.LinkToUserID, i.OAuthIdentitySpec)
		if err != nil {
			return nil, err
		}
		return authflow.NewNodeSimple(&NodeDoCreateIdentity{
			Identity: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentAccountLinkingOAuth) createSyntheticInputOAuthConflict(oauthIden *identity.Spec, conflictedInfo *identity.Info) *SyntheticInputOAuthConflict {
	input := &SyntheticInputOAuthConflict{}

	switch conflictedInfo.Type {
	case model.IdentityTypeLoginID:
		input.LoginID = conflictedInfo.LoginID.LoginID
		switch conflictedInfo.LoginID.LoginIDType {
		case model.LoginIDKeyTypeEmail:
			input.Identification = config.AuthenticationFlowIdentificationEmail
		case model.LoginIDKeyTypePhone:
			input.Identification = config.AuthenticationFlowIdentificationPhone
		case model.LoginIDKeyTypeUsername:
			input.Identification = config.AuthenticationFlowIdentificationUsername
		}
	case model.IdentityTypeOAuth:
		input.Identification = config.AuthenticationFlowIdentificationOAuth
	case model.IdentityTypePasskey:
		input.Identification = config.AuthenticationFlowIdentificationPasskey
	default:
		// This is a panic because the node should not provide option that we don't know how to handle to the user
		panic(fmt.Errorf("unable to create synthetic input from identity type %v", conflictedInfo.Type))
	}
	return input
}
