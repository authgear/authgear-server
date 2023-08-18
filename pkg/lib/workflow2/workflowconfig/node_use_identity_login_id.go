package workflowconfig

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	workflow "github.com/authgear/authgear-server/pkg/lib/workflow2"
)

func init() {
	workflow.RegisterNode(&NodeUseIdentityLoginID{})
}

type NodeUseIdentityLoginID struct {
	Identification config.WorkflowIdentificationMethod `json:"identification,omitempty"`
}

var _ workflow.NodeSimple = &NodeUseIdentityLoginID{}
var _ workflow.Milestone = &NodeUseIdentityLoginID{}
var _ MilestoneIdentificationMethod = &NodeUseIdentityLoginID{}
var _ workflow.InputReactor = &NodeUseIdentityLoginID{}

func (*NodeUseIdentityLoginID) Kind() string {
	return "workflowconfig.NodeUseIdentityLoginID"
}

func (*NodeUseIdentityLoginID) Milestone() {}
func (n *NodeUseIdentityLoginID) MilestoneIdentificationMethod() config.WorkflowIdentificationMethod {
	return n.Identification
}

func (*NodeUseIdentityLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (workflow.InputSchema, error) {
	return &InputTakeLoginID{}, nil
}

func (n *NodeUseIdentityLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if workflow.AsInput(input, &inputTakeLoginID) {
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

		n, err := NewNodeDoUseIdentity(workflows, &NodeDoUseIdentity{
			Identity: exactMatch,
		})
		if err != nil {
			return nil, err
		}

		return workflow.NewNodeSimple(n), nil
	}

	return nil, workflow.ErrIncompatibleInput
}
