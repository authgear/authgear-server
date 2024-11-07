package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentCreateOOBOTPAuthenticatorForLoginID{})
}

var IntentCreateOOBOTPAuthenticatorForLoginIDSchema = validation.NewSimpleSchema(`{}`)

type IntentCreateOOBOTPAuthenticatorForLoginID struct {
	Identity               *identity.Info `json:"identity_info,omitempty"`
	AuthenticatorIsDefault bool           `json:"authenticator_is_default"`
}

var _ NewAuthenticatorGetter = &IntentCreateOOBOTPAuthenticatorForLoginID{}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) Kind() string {
	return "latte.IntentCreateOOBOTPAuthenticatorForLoginID"
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) JSONSchema() *validation.SimpleSchema {
	return IntentCreateOOBOTPAuthenticatorForLoginIDSchema
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentCreateOOBOTPAuthenticatorForLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	channel, target := i.Identity.LoginID.Deprecated_ToChannelTarget()

	authenticatorType, err := model.Deprecated_GetOOBAuthenticatorType(channel)
	if err != nil {
		return nil, err
	}

	spec := &authenticator.Spec{
		Type:      authenticatorType,
		UserID:    i.Identity.UserID,
		IsDefault: i.AuthenticatorIsDefault,
		// It must be primary because it is for a login ID.
		Kind:   authenticator.KindPrimary,
		OOBOTP: &authenticator.OOBOTPSpec{},
	}
	switch channel {
	case model.AuthenticatorOOBChannelSMS:
		spec.OOBOTP.Phone = target
	case model.AuthenticatorOOBChannelEmail:
		spec.OOBOTP.Email = target
	}

	authenticatorID := uuid.New()

	info, err := deps.Authenticators.NewWithAuthenticatorID(ctx, authenticatorID, spec)
	if err != nil {
		return nil, err
	}

	return workflow.NewNodeSimple(&NodeDoCreateAuthenticator{
		Authenticator: info,
	}), nil
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) GetNewAuthenticators(w *workflow.Workflow) ([]*authenticator.Info, bool) {
	node, ok := workflow.FindSingleNode[*NodeDoCreateAuthenticator](w)
	if !ok {
		return nil, false
	}
	return []*authenticator.Info{node.Authenticator}, true
}
