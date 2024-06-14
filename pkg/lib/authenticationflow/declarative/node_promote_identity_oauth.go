package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

func init() {
	authflow.RegisterNode(&NodePromoteIdentityOAuth{})
}

type NodePromoteIdentityOAuth struct {
	JSONPointer    jsonpointer.T      `json:"json_pointer,omitempty"`
	UserID         string             `json:"user_id,omitempty"`
	SyntheticInput *InputStepIdentify `json:"synthetic_input,omitempty"`
	Alias          string             `json:"alias,omitempty"`
	RedirectURI    string             `json:"redirect_uri,omitempty"`
	ResponseMode   string             `json:"response_mode,omitempty"`
}

var _ authflow.NodeSimple = &NodePromoteIdentityOAuth{}
var _ authflow.InputReactor = &NodePromoteIdentityOAuth{}
var _ authflow.DataOutputer = &NodePromoteIdentityOAuth{}

func (*NodePromoteIdentityOAuth) Kind() string {
	return "NodePromoteIdentityOAuth"
}

func (n *NodePromoteIdentityOAuth) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}
	return &InputSchemaTakeOAuthAuthorizationResponse{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (n *NodePromoteIdentityOAuth) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputOAuth inputTakeOAuthAuthorizationResponse
	if authflow.AsInput(input, &inputOAuth) {
		spec, err := handleOAuthAuthorizationResponse(deps, HandleOAuthAuthorizationResponseOptions{
			Alias:       n.Alias,
			RedirectURI: n.RedirectURI,
		}, inputOAuth)
		if err != nil {
			return nil, err
		}

		syntheticInput := &SyntheticInputOAuth{
			Identification: n.SyntheticInput.Identification,
			Alias:          n.SyntheticInput.Alias,
			RedirectURI:    n.SyntheticInput.RedirectURI,
			ResponseMode:   n.SyntheticInput.ResponseMode,
			IdentitySpec:   spec,
		}

		_, err = findExactOneIdentityInfo(deps, spec)
		if err != nil {
			if apierrors.IsKind(err, api.UserNotFound) {
				conflicts, err := n.checkConflictByAccountLinkings(ctx, deps, flows, spec)
				if err != nil {
					return nil, err
				}
				if len(conflicts) > 0 {
					// In promote flow, always error if any conflicts occurs
					conflictSpecs := slice.Map(conflicts, func(c *AccountLinkingConflict) *identity.Spec {
						s := c.Identity.ToSpec()
						return &s
					})
					return nil, identityFillDetailsMany(api.ErrDuplicatedIdentity, spec, conflictSpecs)
				}

				// promote
				info, err := newIdentityInfo(deps, n.UserID, spec)
				if err != nil {
					return nil, err
				}

				return authflow.NewNodeSimple(&NodeDoCreateIdentity{
					Identity: info,
				}), nil
			}
			// general error
			return nil, err
		}

		// login
		flowReference := authflow.FindCurrentFlowReference(flows.Root)
		return nil, &authflow.ErrorSwitchFlow{
			FlowReference: authflow.FlowReference{
				Type: authflow.FlowTypeLogin,
				// Switch to the login flow of the same name.
				Name: flowReference.Name,
			},
			SyntheticInput: syntheticInput,
		}
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodePromoteIdentityOAuth) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
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

func (n *NodePromoteIdentityOAuth) checkConflictByAccountLinkings(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows,
	spec *identity.Spec) (conflicts []*AccountLinkingConflict, err error) {
	switch spec.Type {
	case model.IdentityTypeOAuth:
		return linkByIncomingOAuthSpec(ctx, deps, flows, NewCreateOAuthIdentityRequest(n.Alias, spec).OAuth, n.JSONPointer)
	default:
		panic("unexpected spec type")
	}
}
