package workflowconfig

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

func authenticatorIsDefault(deps *workflow.Dependencies, userID string, authenticatorKind model.AuthenticatorKind) (isDefault bool, err error) {
	ais, err := deps.Authenticators.List(
		userID,
		authenticator.KeepKind(authenticatorKind),
		authenticator.KeepDefault,
	)
	if err != nil {
		return
	}

	isDefault = len(ais) == 0
	return
}

func findSignupFlow(c *config.WorkflowConfig, id string) (config.WorkflowObject, error) {
	for _, f := range c.SignupFlows {
		if f.ID == id {
			f := f
			return f, nil
		}
	}
	return nil, ErrFlowNotFound
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
