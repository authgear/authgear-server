package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeDoFillForm{})
}

type NodeDoFillForm struct {
	UserID string `json:"user_id,omitempty"`
	// JSONPointer points to the flow object that has the form.
	JSONPointer jsonpointer.T    `json:"json_pointer,omitempty"`
	Fields      []FormFieldInput `json:"fields,omitempty"`
	SkipUpdate  bool             `json:"skip_update,omitempty"`
}

var _ authflow.NodeSimple = &NodeDoFillForm{}
var _ authflow.EffectGetter = &NodeDoFillForm{}
var _ authflow.Milestone = &NodeDoFillForm{}
var _ MilestoneDidFillForm = &NodeDoFillForm{}
var _ MilestoneSwitchToExistingUser = &NodeDoFillForm{}

func (*NodeDoFillForm) Kind() string {
	return "NodeDoFillForm"
}

func (*NodeDoFillForm) Milestone() {}

func (n *NodeDoFillForm) MilestoneDidFillForm() (skipped bool) {
	return n.SkipUpdate
}

// MilestoneSwitchToExistingUser is called when account linking rewrites the flow to an existing user
// after the form was submitted. Forms are skipped then, and replaying the effect
// against a user ID that is never created would break every later request.
func (n *NodeDoFillForm) MilestoneSwitchToExistingUser(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error {
	n.SkipUpdate = true
	return nil
}

func (n *NodeDoFillForm) GetEffects(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (effs []authflow.Effect, err error) {
	return []authflow.Effect{
		authflow.RunEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
			if n.SkipUpdate {
				return nil
			}

			rootObject, err := findNearestFlowObjectInFlow(deps, flows, n)
			if err != nil {
				return err
			}
			current, err := authflow.FlowObject(rootObject, n.JSONPointer)
			if err != nil {
				return err
			}
			form := formFromFlowObject(current)

			std, custom := formFieldsToAttrs(deps.Config, form, n.Fields)

			// No event is dispatched here.
			// user.profile.pre_update and user.profile.updated are dispatched once when the flow finishes.
			// FIXME(authflow): support other role?
			if len(std) > 0 {
				err = deps.StdAttrsService.UpdateStandardAttributesWithList(ctx, config.RoleEndUser, n.UserID, std)
				if err != nil {
					return err
				}
			}
			if len(custom) > 0 {
				err = deps.CustomAttrsService.UpdateCustomAttributesWithList(ctx, config.RoleEndUser, n.UserID, custom)
				if err != nil {
					return err
				}
			}
			return nil
		}),
	}, nil
}
