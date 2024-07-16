package latte

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentCreateLoginID{})
}

var IntentCreateLoginIDSchema = validation.NewSimpleSchema(`{}`)

type IntentCreateLoginID struct {
	CaptchaProtectedIntent
	UserID      string               `json:"user_id"`
	LoginIDType model.LoginIDKeyType `json:"login_id_type"`
	LoginIDKey  string               `json:"login_id_key"`
}

var _ NewIdentityGetter = &IntentCreateLoginID{}
var _ NewAuthenticatorGetter = &IntentCreateLoginID{}

func (*IntentCreateLoginID) Kind() string {
	return "latte.IntentCreateLoginID"
}

func (*IntentCreateLoginID) JSONSchema() *validation.SimpleSchema {
	return IntentCreateLoginIDSchema
}

func (*IntentCreateLoginID) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return []workflow.Input{
			&InputTakeLoginID{},
		}, nil
	case 1:
		return nil, nil
	case 2:
		return nil, nil
	case 3:
		return nil, nil
	default:
		return nil, workflow.ErrEOF
	}
}

func (i *IntentCreateLoginID) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID

	switch len(workflows.Nearest.Nodes) {
	case 0:
		if workflow.AsInput(input, &inputTakeLoginID) {
			loginID := inputTakeLoginID.GetLoginID()
			spec := &identity.Spec{
				Type: model.IdentityTypeLoginID,
				LoginID: &identity.LoginIDSpec{
					Type:  i.LoginIDType,
					Key:   i.LoginIDKey,
					Value: loginID,
				},
			}

			info, err := deps.Identities.New(i.UserID, spec, identity.NewIdentityOptions{
				LoginIDEmailByPassBlocklistAllowlist: false,
			})
			if err != nil {
				return nil, err
			}

			_, err = deps.Identities.CheckDuplicated(info)
			if err != nil {
				return nil, err
			}

			return workflow.NewNodeSimple(&NodeDoCreateIdentity{
				Identity: info,
			}), nil
		}
	case 1:
		iden := i.identityInfo(workflows.Nearest)
		return workflow.NewNodeSimple(&NodePopulateStandardAttributes{
			Identity: iden,
		}), nil
	case 2:
		iden := i.identityInfo(workflows.Nearest)
		intent := &IntentVerifyIdentity{
			Identity:     iden,
			IsFromSignUp: true,
		}
		intent.IsCaptchaProtected = i.IsCaptchaProtected
		return workflow.NewSubWorkflow(intent), nil
	case 3:
		iden := i.identityInfo(workflows.Nearest)
		return workflow.NewSubWorkflow(&IntentCreateOOBOTPAuthenticatorForLoginID{
			Identity: iden,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentCreateLoginID) identityInfo(w *workflow.Workflow) *identity.Info {
	node, ok := workflow.FindSingleNode[*NodeDoCreateIdentity](w)
	if !ok {
		panic(fmt.Errorf("workflow: expected NodeCreateIdentity"))
	}

	return node.Identity
}

func (*IntentCreateLoginID) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return nil, nil
}

func (*IntentCreateLoginID) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (*IntentCreateLoginID) GetNewIdentities(w *workflow.Workflow) ([]*identity.Info, bool) {
	node, ok := workflow.FindSingleNode[*NodeDoCreateIdentity](w)
	if !ok {
		return nil, false
	}
	return []*identity.Info{node.Identity}, true
}

func (*IntentCreateLoginID) GetNewAuthenticators(w *workflow.Workflow) ([]*authenticator.Info, bool) {
	ws := workflow.FindSubWorkflows[NewAuthenticatorGetter](w)
	if len(ws) != 1 {
		return nil, false
	}
	subworkflow := ws[0]
	return subworkflow.Intent.(NewAuthenticatorGetter).GetNewAuthenticators(subworkflow)
}

func identityFillDetails(err error, spec *identity.Spec, otherSpec *identity.Spec) error {
	details := errorutil.Details{}

	if spec != nil {
		details["IdentityTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.Type)
		switch spec.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeIncoming"] = apierrors.APIErrorDetail.Value(spec.OAuth.ProviderID.Type)
		}
	}

	if otherSpec != nil {
		details["IdentityTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.Type)
		switch otherSpec.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeExisting"] = apierrors.APIErrorDetail.Value(otherSpec.OAuth.ProviderID.Type)
		}
	}

	return errorutil.WithDetails(err, details)
}
