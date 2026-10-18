package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterIntent(&IntentFillForm{})
}

// IntentFillForm shows a form, writes the submitted values, and dispatches authentication.form.post_submitted.
//
//	IntentFillForm (MilestoneFillForm)
//	  NodeDoFillForm (MilestoneDidFillForm)
//	  NodeFormPostSubmitted
type IntentFillForm struct {
	// JSONPointer points to a fill_form step, or an identify branch that has a form.
	JSONPointer            jsonpointer.T `json:"json_pointer,omitempty"`
	UserID                 string        `json:"user_id,omitempty"`
	IsUpdatingExistingUser bool          `json:"is_updating_existing_user,omitempty"`
	// IsSubmittedWithIdentification is true when the values came with the login ID in an identify step.
	IsSubmittedWithIdentification bool `json:"is_submitted_with_identification,omitempty"`
	// SubmittedFields is only held until the first ReactTo, which hands it to NodeDoFillForm.
	SubmittedFields []FormFieldInput `json:"submitted_fields,omitempty"`
}

var _ authflow.Intent = &IntentFillForm{}
var _ authflow.DataOutputer = &IntentFillForm{}
var _ authflow.Milestone = &IntentFillForm{}
var _ MilestoneFillForm = &IntentFillForm{}
var _ MilestoneSwitchToExistingUser = &IntentFillForm{}

func (*IntentFillForm) Kind() string {
	return "IntentFillForm"
}

func (*IntentFillForm) Milestone()         {}
func (*IntentFillForm) MilestoneFillForm() {}
func (i *IntentFillForm) MilestoneSwitchToExistingUser(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, newUserID string) error {
	i.UserID = newUserID
	i.IsUpdatingExistingUser = true
	return nil
}

func (i *IntentFillForm) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		if i.IsUpdatingExistingUser {
			// Forms are skipped when the flow continues as an existing user.
			return nil, authflow.ErrEOF
		}
		if i.IsSubmittedWithIdentification {
			return nil, nil
		}

		rootObject, form, err := i.rootObjectAndForm(deps, flows)
		if err != nil {
			return nil, err
		}
		return &InputSchemaFillForm{
			JSONPointer:    i.JSONPointer,
			FlowRootObject: rootObject,
			Config:         deps.Config,
			Form:           form,
		}, nil
	case 1:
		return nil, nil
	default:
		return nil, authflow.ErrEOF
	}
}

func (i *IntentFillForm) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	switch len(flows.Nearest.Nodes) {
	case 0:
		var fields []FormFieldInput
		if i.IsSubmittedWithIdentification {
			fields = i.SubmittedFields
		} else {
			// One raw input feeds at most one form.
			if authflow.IsFormInputConsumed(ctx) {
				return nil, authflow.ErrIncompatibleInput
			}
			var inputFillForm inputFillForm
			if !authflow.AsInput(input, &inputFillForm) {
				return nil, authflow.ErrIncompatibleInput
			}
			fields = inputFillForm.GetFormFields()
		}
		// The prefilled path also marks it, so a later fill_form never consumes the identify input.
		authflow.MarkFormInputConsumed(ctx)

		// The node holds the only copy from now on.
		i.SubmittedFields = nil

		return authflow.NewNodeSimple(&NodeDoFillForm{
			UserID:      i.UserID,
			JSONPointer: i.JSONPointer,
			Fields:      fields,
		}), nil
	case 1:
		node, ok := flows.Nearest.Nodes[0].Simple.(*NodeDoFillForm)
		if !ok {
			return nil, authflow.ErrIncompatibleInput
		}
		_, form, err := i.rootObjectAndForm(deps, flows)
		if err != nil {
			return nil, err
		}
		return NewNodeFormPostSubmitted(ctx, deps, flows, &NodeFormPostSubmittedOptions{
			UserID: i.UserID,
			Form:   form,
			Fields: node.Fields,
		})
	default:
		return nil, authflow.ErrIncompatibleInput
	}
}

func (i *IntentFillForm) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	_, form, err := i.rootObjectAndForm(deps, flows)
	if err != nil {
		return nil, err
	}

	// FIXME(authflow): support other role?
	user, err := deps.Users.Get(ctx, i.UserID, config.RoleEndUser)
	if err != nil {
		return nil, err
	}

	dataForm, err := NewFillFormDataForm(ctx, deps, form, user)
	if err != nil {
		return nil, err
	}
	return NewFillFormData(dataForm), nil
}

func (i *IntentFillForm) rootObjectAndForm(deps *authflow.Dependencies, flows authflow.Flows) (config.AuthenticationFlowObject, *config.AuthenticationFlowForm, error) {
	rootObject, err := findNearestFlowObjectInFlow(deps, flows, i)
	if err != nil {
		return nil, nil, err
	}
	current, err := authflow.FlowObject(rootObject, i.JSONPointer)
	if err != nil {
		return nil, nil, err
	}
	return rootObject, formFromFlowObject(current), nil
}
