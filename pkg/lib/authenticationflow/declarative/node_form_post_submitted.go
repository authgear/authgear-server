package declarative

import (
	"context"

	eventapi "github.com/authgear/authgear-server/pkg/api/event"
	blocking "github.com/authgear/authgear-server/pkg/api/event/blocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeFormPostSubmitted{})
}

type NodeFormPostSubmittedOptions struct {
	UserID string
	Form   *config.AuthenticationFlowForm
	Fields []FormFieldInput
}

// NewNodeFormPostSubmitted prepares the authentication.form.post_submitted event.
// The user in the payload is resolved here, inside the transaction where NodeDoFillForm has been applied,
// so it has every attribute submitted so far.
func NewNodeFormPostSubmitted(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, opts *NodeFormPostSubmittedOptions) (authflow.ReactToResult, error) {
	n := &NodeFormPostSubmitted{}

	authCtx, err := GetAuthenticationContext(ctx, deps, flows)
	if err != nil {
		return nil, err
	}

	payload := &blocking.AuthenticationFormPostSubmittedBlockingEventPayload{
		UserRef:               model.UserRef{Meta: model.Meta{ID: opts.UserID}},
		AuthenticationContext: *authCtx,
		Form:                  makeSubmittedForm(opts.Form, opts.Fields),
	}
	e, err := deps.Events.PrepareBlockingEventWithTx(ctx, payload, eventapi.PrepareBlockingEventOptions{})
	if err != nil {
		return nil, err
	}

	var delayedFunction authflow.DelayedOneTimeFunction = func(ctx context.Context, deps *authflow.Dependencies) (authflow.DelayedOneTimeFunctionResult, error) {
		err := deps.Events.DispatchEventWithoutTx(ctx, e)
		if err != nil {
			return authflow.DelayedOneTimeFunctionResult{}, err
		}
		n.IsInvoked = true
		return authflow.DelayedOneTimeFunctionResult{}, nil
	}

	return &authflow.NodeReactToResult{
		Node:                   authflow.NewNodeSimple(n),
		DelayedOneTimeFunction: delayedFunction,
	}, nil
}

// makeSubmittedForm lists every field of the form in order.
// Value is nil when the field is absent from the input.
func makeSubmittedForm(form *config.AuthenticationFlowForm, fields []FormFieldInput) eventapi.SubmittedForm {
	submitted := map[string]any{}
	for _, f := range fields {
		submitted[f.Key] = f.Value
	}

	out := eventapi.SubmittedForm{
		Name:   form.Name,
		Fields: []eventapi.SubmittedFormField{},
	}
	for _, f := range form.Fields {
		key := formFieldKey(f)
		out.Fields = append(out.Fields, eventapi.SubmittedFormField{
			Field: eventapi.FormFieldRef{UserProfile: &eventapi.FormFieldRefUserProfile{Pointer: key}},
			Value: submitted[key],
		})
	}
	return out
}

type NodeFormPostSubmitted struct {
	IsInvoked bool `json:"is_invoked"`
}

var _ authflow.NodeSimple = &NodeFormPostSubmitted{}
var _ authflow.InputReactor = &NodeFormPostSubmitted{}

func (*NodeFormPostSubmitted) Kind() string {
	return "NodeFormPostSubmitted"
}

func (n *NodeFormPostSubmitted) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	if n.IsInvoked {
		return nil, authflow.ErrEOF
	}
	return nil, authflow.ErrPauseAndRetryAccept
}

func (n *NodeFormPostSubmitted) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (authflow.ReactToResult, error) {
	return nil, authflow.ErrEOF
}
