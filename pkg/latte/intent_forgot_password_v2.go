package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/stringutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentForgotPasswordV2{})
}

var IntentForgotPasswordV2Schema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentForgotPasswordV2 struct {
}

func (*IntentForgotPasswordV2) Kind() string {
	return "latte.IntentForgotPasswordV2"
}

func (*IntentForgotPasswordV2) JSONSchema() *validation.SimpleSchema {
	return IntentForgotPasswordSchema
}

func (*IntentForgotPasswordV2) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {

	switch len(workflows.Nearest.Nodes) {
	case 0:
		return []workflow.Input{
			&InputTakeLoginID{},
		}, nil
	case 1:
		return []workflow.Input{
			&InputTakeForgotPasswordChannel{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentForgotPasswordV2) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		var inputTakeLoginID inputTakeLoginID
		if workflow.AsInput(input, &inputTakeLoginID) {
			loginID := inputTakeLoginID.GetLoginID()
			spec := &identity.Spec{
				Type: model.IdentityTypeLoginID,
				LoginID: &identity.LoginIDSpec{
					Value: stringutil.NewUserInputString(loginID),
				},
			}

			exactMatch, _, err := deps.Identities.SearchBySpec(ctx, spec)
			if err != nil {
				return nil, err
			}

			var userID *string
			if exactMatch != nil {
				userID = &exactMatch.UserID
			}

			return workflow.NewNodeSimple(&NodeForgotPasswordForUser{
				LoginID: loginID,
				UserID:  userID,
			}), nil
		}
	case 1:
		var inputTakeForgotPasswordChannel inputTakeForgotPasswordChannel
		if workflow.AsInput(input, &inputTakeForgotPasswordChannel) {
			channel := inputTakeForgotPasswordChannel.GetForgotPasswordChannel()
			node, err := i.sendCodeForChannel(ctx, workflows.Nearest, deps, channel)
			if err != nil {
				return nil, err
			}
			return workflow.NewNodeSimple(node), nil
		}
	}
	return nil, workflow.ErrIncompatibleInput
}

func (*IntentForgotPasswordV2) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentForgotPasswordV2) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentForgotPasswordV2) sendCodeForChannel(
	ctx context.Context,
	w *workflow.Workflow,
	deps *workflow.Dependencies,
	channel ForgotPasswordChannel) (*NodeSendForgotPasswordCode, error) {
	prevnode, found := workflow.FindSingleNode[*NodeForgotPasswordForUser](w)
	if !found {
		panic("NodeForgotPasswordForUser not found but it must exist")
	}
	if prevnode.UserID == nil {
		return &NodeSendForgotPasswordCode{LoginID: prevnode.LoginID}, nil
	}

	targetLoginID, err := selectForgotPasswordLoginID(ctx, deps, *prevnode.UserID, channel)

	if err == nil || errors.Is(err, ErrNoMatchingLoginIDForForgotPasswordChannel) {

		if targetLoginID != "" {
			err = deps.ForgotPassword.SendCode(ctx, targetLoginID, nil)
			if err != nil {
				return nil, err
			}
		}

		return &NodeSendForgotPasswordCode{LoginID: prevnode.LoginID}, nil
	}

	return nil, err
}
