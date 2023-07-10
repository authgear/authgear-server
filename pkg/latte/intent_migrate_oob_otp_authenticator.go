package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentMigrateOOBOTPAuthenticator{})
}

var IntentMigrateOOBOTPAuthenticatorSchema = validation.NewSimpleSchema(`{}`)

type IntentMigrateOOBOTPAuthenticator struct {
	UserID                 string                     `json:"user_id"`
	MigrateSpec            *authenticator.MigrateSpec `json:"migrate_spec"`
	AuthenticatorIsDefault bool                       `json:"authenticator_is_default"`
}

func (*IntentMigrateOOBOTPAuthenticator) Kind() string {
	return "latte.IntentMigrateOOBOTPAuthenticator"
}

func (*IntentMigrateOOBOTPAuthenticator) JSONSchema() *validation.SimpleSchema {
	return IntentMigrateOOBOTPAuthenticatorSchema
}

func (*IntentMigrateOOBOTPAuthenticator) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		return nil, nil
	}
	return nil, workflow.ErrEOF
}

func (i *IntentMigrateOOBOTPAuthenticator) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	spec := i.MigrateSpec.GetSpec()
	spec.UserID = i.UserID
	spec.IsDefault = i.AuthenticatorIsDefault

	// Normalize the target.
	switch spec.Type {
	case model.AuthenticatorTypeOOBEmail:
		email := spec.OOBOTP.Email
		var err error
		email, err = deps.LoginIDNormalizerFactory.NormalizerWithLoginIDType(model.LoginIDKeyTypeEmail).Normalize(email)
		if err != nil {
			return nil, err
		}
		spec.OOBOTP.Email = email
	case model.AuthenticatorTypeOOBSMS:
		phone := spec.OOBOTP.Phone
		var err error
		phone, err = deps.LoginIDNormalizerFactory.NormalizerWithLoginIDType(model.LoginIDKeyTypePhone).Normalize(phone)
		if err != nil {
			return nil, err
		}
		spec.OOBOTP.Phone = phone
	default:
		panic("workflow: creating OOB authenticator for invalid channel")
	}

	authenticatorID := uuid.New()
	info, err := deps.Authenticators.NewWithAuthenticatorID(authenticatorID, spec)
	if err != nil {
		return nil, err
	}

	return workflow.NewNodeSimple(&NodeDoCreateAuthenticator{
		Authenticator: info,
	}), nil
}

func (*IntentMigrateOOBOTPAuthenticator) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentMigrateOOBOTPAuthenticator) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentMigrateOOBOTPAuthenticator) GetNewAuthenticators(w *workflow.Workflow) ([]*authenticator.Info, bool) {
	node, ok := workflow.FindSingleNode[*NodeDoCreateAuthenticator](w)
	if !ok {
		return nil, false
	}
	return []*authenticator.Info{node.Authenticator}, true
}
