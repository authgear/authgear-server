package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeUseIdentityLoginID{})
}

type NodeUseIdentityLoginID struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.NodeSimple = &NodeUseIdentityLoginID{}
var _ authflow.Milestone = &NodeUseIdentityLoginID{}
var _ MilestoneIdentificationMethod = &NodeUseIdentityLoginID{}
var _ authflow.InputReactor = &NodeUseIdentityLoginID{}

func (*NodeUseIdentityLoginID) Kind() string {
	return "NodeUseIdentityLoginID"
}

func (*NodeUseIdentityLoginID) Milestone() {}
func (n *NodeUseIdentityLoginID) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *NodeUseIdentityLoginID) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaTakeLoginID{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeUseIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		bucketSpec := AccountEnumerationPerIPRateLimitBucketSpec(
			deps.Config.Authentication,
			string(deps.RemoteIP),
		)

		reservation := deps.RateLimiter.Reserve(bucketSpec)
		err := reservation.Error()
		if err != nil {
			return nil, err
		}
		defer deps.RateLimiter.Cancel(reservation)

		loginID := inputTakeLoginID.GetLoginID()
		spec := &identity.Spec{
			Type: model.IdentityTypeLoginID,
			LoginID: &identity.LoginIDSpec{
				Value: loginID,
			},
		}

		exactMatch, otherMatches, err := deps.Identities.SearchBySpec(spec)
		if err != nil {
			return nil, err
		}

		if exactMatch == nil {
			// Consume the reservation if exact match is not found.
			reservation.Consume()

			var otherSpec *identity.Spec
			if len(otherMatches) > 0 {
				s := otherMatches[0].ToSpec()
				otherSpec = &s
			}
			return nil, identityFillDetails(api.ErrUserNotFound, spec, otherSpec)
		}

		n, err := NewNodeDoUseIdentity(flows, &NodeDoUseIdentity{
			Identity: exactMatch,
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(n), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
