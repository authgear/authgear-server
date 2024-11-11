package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentReauthForgotPassword{})
}

var IntentReauthForgotPasswordSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentReauthForgotPassword struct {
}

func (*IntentReauthForgotPassword) Kind() string {
	return "latte.IntentReauthForgotPassword"
}

func (*IntentReauthForgotPassword) JSONSchema() *validation.SimpleSchema {
	return IntentForgotPasswordSchema
}

func (*IntentReauthForgotPassword) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {

	switch len(workflows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return []workflow.Input{
			&InputTakeForgotPasswordChannel{},
		}, nil
	case 2:
		return []workflow.Input{
			&InputSendForgotPasswordCode{},
		}, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentReauthForgotPassword) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	userID, err := reauthUserIDHint(ctx)
	if err != nil {
		return nil, err
	}

	switch len(workflows.Nearest.Nodes) {
	case 0:
		return workflow.NewNodeSimple(&NodeForgotPasswordForUser{
			LoginID: "",
			UserID:  &userID,
		}), nil
	case 1:
		var inputTakeForgotPasswordChannel inputTakeForgotPasswordChannel
		if workflow.AsInput(input, &inputTakeForgotPasswordChannel) {
			channel := inputTakeForgotPasswordChannel.GetForgotPasswordChannel()
			node, err := i.selectLoginIDForChannel(ctx, workflows.Nearest, deps, channel)
			if err != nil {
				return nil, err
			}
			return workflow.NewNodeSimple(node), nil
		}
	case 2:
		var inputSendForgotPasswordCode inputSendForgotPasswordCode
		if workflow.AsInput(input, &inputSendForgotPasswordCode) {
			node, err := i.sendCode(ctx, workflows.Nearest, deps)
			if err != nil {
				return nil, err
			}
			return workflow.NewNodeSimple(node), nil
		}

	}
	return nil, workflow.ErrIncompatibleInput
}

func (*IntentReauthForgotPassword) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentReauthForgotPassword) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentReauthForgotPassword) selectLoginIDForChannel(
	ctx context.Context,
	w *workflow.Workflow,
	deps *workflow.Dependencies,
	channel ForgotPasswordChannel) (*NodeForgotPasswordWithLoginID, error) {
	prevnode, found := workflow.FindSingleNode[*NodeForgotPasswordForUser](w)
	if !found {
		panic("NodeForgotPasswordForUser not found but it must exist")
	}
	if prevnode.UserID == nil {
		panic("UserID is nil in NodeForgotPasswordForUser but it must not be nil")
	}

	targetLoginID, err := selectForgotPasswordLoginID(ctx, deps, *prevnode.UserID, channel)

	if err != nil {
		return nil, err
	}

	return &NodeForgotPasswordWithLoginID{LoginID: targetLoginID}, nil
}

func (*IntentReauthForgotPassword) sendCode(
	ctx context.Context,
	w *workflow.Workflow,
	deps *workflow.Dependencies) (*NodeSendForgotPasswordCode, error) {
	prevnode, found := workflow.FindSingleNode[*NodeForgotPasswordWithLoginID](w)
	if !found {
		panic("NodeForgotPasswordWithLoginID not found but it must exist")
	}
	loginID := prevnode.LoginID

	newNode := &NodeSendForgotPasswordCode{
		LoginID: loginID,
	}

	err := newNode.sendCode(ctx, deps)
	if err != nil {
		return nil, err
	}

	return newNode, nil
}
