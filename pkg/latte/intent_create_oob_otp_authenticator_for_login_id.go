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
	IdentityInfo           *identity.Info `json:"identity_info,omitempty"`
	AuthenticatorIsDefault bool           `json:"authenticator_is_default"`
}

var _ NewAuthenticatorGetter = &IntentCreateOOBOTPAuthenticatorForLoginID{}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) Kind() string {
	return "latte.IntentCreateOOBOTPAuthenticatorForLoginID"
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) JSONSchema() *validation.SimpleSchema {
	return IntentCreateOOBOTPAuthenticatorForLoginIDSchema
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	if len(w.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentCreateOOBOTPAuthenticatorForLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	channel, target := i.IdentityInfo.LoginID.ToChannelTarget()

	authenticatorType, err := model.GetOOBAuthenticatorType(channel)
	if err != nil {
		return nil, err
	}

	// Validate target against channel
	validationCtx := &validation.Context{}
	switch channel {
	case model.AuthenticatorOOBChannelEmail:
		err := validation.FormatEmail{AllowName: false}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "email"})
		}
	case model.AuthenticatorOOBChannelSMS:
		err := validation.FormatPhone{}.CheckFormat(target)
		if err != nil {
			validationCtx.EmitError("format", map[string]interface{}{"format": "phone"})
		}
	}
	err = validationCtx.Error("invalid target")
	if err != nil {
		return nil, err
	}

	spec := &authenticator.Spec{
		Type:      authenticatorType,
		UserID:    i.IdentityInfo.UserID,
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

	info, err := deps.Authenticators.NewWithAuthenticatorID(authenticatorID, spec)
	if err != nil {
		return nil, err
	}

	return workflow.NewNodeSimple(&NodeDoCreateAuthenticator{
		AuthenticatorInfo: info,
	}), nil
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (*IntentCreateOOBOTPAuthenticatorForLoginID) GetNewAuthenticator(w *workflow.Workflow) (*authenticator.Info, bool) {
	node, ok := workflow.FindSingleNode[*NodeDoCreateAuthenticator](w)
	if !ok {
		return nil, false
	}
	return node.AuthenticatorInfo, true
}
