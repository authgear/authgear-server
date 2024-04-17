package declarative

import (
	"context"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/authn/sso"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeOAuth{})
}

type NodeOAuth struct {
	JSONPointer  jsonpointer.T    `json:"json_pointer,omitempty"`
	NewUserID    string           `json:"new_user_id,omitempty"`
	Alias        string           `json:"alias,omitempty"`
	RedirectURI  string           `json:"redirect_uri,omitempty"`
	ResponseMode sso.ResponseMode `json:"response_mode,omitempty"`
}

var _ authflow.NodeSimple = &NodeOAuth{}
var _ authflow.InputReactor = &NodeOAuth{}
var _ authflow.DataOutputer = &NodeOAuth{}

func (*NodeOAuth) Kind() string {
	return "NodeOAuth"
}

func (n *NodeOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaTakeOAuthAuthorizationResponse{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var syntheticInputOAuth syntheticInputOAuth
	var inputOAuth inputTakeOAuthAuthorizationResponse
	// The order of the cases is important.
	// We must handle the synthetic input first.
	// It is because if it is synthetic input,
	// then the code has been consumed.
	// Using the code again will definitely fail.
	switch {
	case authflow.AsInput(input, &syntheticInputOAuth):
		spec := syntheticInputOAuth.GetIdentitySpec()
		return n.reactTo(ctx, deps, flows, spec)
	case authflow.AsInput(input, &inputOAuth):
		spec, err := handleOAuthAuthorizationResponse(deps, HandleOAuthAuthorizationResponseOptions{
			Alias:       n.Alias,
			RedirectURI: n.RedirectURI,
		}, inputOAuth)
		if err != nil {
			return nil, err
		}

		return n.reactTo(ctx, deps, flows, spec)
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeOAuth) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	data, err := getOAuthData(ctx, deps, GetOAuthDataOptions{
		RedirectURI:  n.RedirectURI,
		Alias:        n.Alias,
		ResponseMode: n.ResponseMode,
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (n *NodeOAuth) reactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, spec *identity.Spec) (*authflow.Node, error) {
	// signup
	if n.NewUserID != "" {
		info, conflictedInfo, err := newIdentityInfo(deps, n.NewUserID, spec)
		if apierrors.IsAPIError(err) && apierrors.AsAPIError(err).HasCause("DuplicatedIdentity") {
			conflictedUserID := conflictedInfo.UserID
			flowUserID, err := getUserID(flows)
			if err != nil {
				return nil, err
			}
			if flowUserID != conflictedUserID {
				authflow.TraverseFlow(authflow.Traverser{
					NodeSimple: func(nodeSimple authflow.NodeSimple, w *authflow.Flow) error {
						milestone, ok := nodeSimple.(MilestoneSwitchToUser)
						if ok {
							milestone.MilestoneSwitchToUser(conflictedUserID)
						}
						return nil
					},
					Intent: func(intent authflow.Intent, w *authflow.Flow) error {
						milestone, ok := intent.(MilestoneSwitchToUser)
						if ok {
							milestone.MilestoneSwitchToUser(conflictedUserID)
						}
						return nil
					},
				}, flows.Root)
				// Use synthetic input to:
				// 1. pass this node in the rewritten flow
				// 2. auto select the conflicted identity in the login flow
				return nil, &authflow.ErrorRewriteFlow{
					Intent:         flows.Root.Intent,
					Nodes:          flows.Root.Nodes,
					SyntheticInput: n.createSyntheticInputOAuthConflict(spec, conflictedInfo),
				}
			}
			// TODO(tung): Check the config to decide what to do
			flowReference := authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				// FIXME(tung): This should be read from config
				Name: "default",
			}
			loginIntent := IntentLoginFlow{
				FlowReference: flowReference,
			}
			return authflow.NewSubFlow(&loginIntent), nil
		}
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoCreateIdentity{
			Identity: info,
		}), nil
	}
	// Else login

	exactMatch, err := findExactOneIdentityInfo(deps, spec)
	if err != nil {
		return nil, err
	}

	newNode, err := NewNodeDoUseIdentity(ctx, flows, &NodeDoUseIdentity{
		Identity: exactMatch,
	})
	if err != nil {
		return nil, err
	}

	return authflow.NewNodeSimple(newNode), nil
}

func (n *NodeOAuth) createSyntheticInputOAuthConflict(oauthSpec *identity.Spec, conflictedInfo *identity.Info) *SyntheticInputOAuthConflict {
	input := &SyntheticInputOAuthConflict{
		OAuthIdentitySpec: oauthSpec,
	}

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
