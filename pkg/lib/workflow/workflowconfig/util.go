package workflowconfig

import (
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

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

func signupFlowCurrent(deps *workflow.Dependencies, id string, pointer jsonpointer.T) (config.WorkflowObject, error) {
	var root config.WorkflowObject
	for _, f := range deps.Config.Workflow.SignupFlows {
		f := f
		if f.ID == id {
			root = f
			break
		}
	}
	if root == nil {
		return nil, ErrFlowNotFound
	}

	entries, err := Traverse(root, pointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func loginFlowCurrent(deps *workflow.Dependencies, id string, pointer jsonpointer.T) (config.WorkflowObject, error) {
	var root config.WorkflowObject
	for _, f := range deps.Config.Workflow.LoginFlows {
		f := f
		if f.ID == id {
			root = f
			break
		}
	}
	if root == nil {
		return nil, ErrFlowNotFound
	}

	entries, err := Traverse(root, pointer)
	if err != nil {
		return nil, err
	}

	current, err := GetCurrentObject(entries)
	if err != nil {
		return nil, err
	}

	return current, nil
}

func getUserID(workflows workflow.Workflows) (userID string, err error) {
	err = workflows.Root.Traverse(workflow.WorkflowTraverser{
		NodeSimple: func(nodeSimple workflow.NodeSimple, w *workflow.Workflow) error {
			if n, ok := nodeSimple.(UserIDGetter); ok {
				id := n.GetUserID()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
		Intent: func(intent workflow.Intent, w *workflow.Workflow) error {
			if i, ok := intent.(UserIDGetter); ok {
				id := i.GetUserID()
				if userID == "" {
					userID = id
				} else if userID != "" && id != userID {
					return ErrDifferentUserID
				}
			}
			return nil
		},
	})

	if userID == "" {
		err = ErrNoUserID
	}

	if err != nil {
		return
	}

	return
}

func getAuthenticationMethodsOfUser(deps *workflow.Dependencies, userID string, allAllowed []config.WorkflowAuthenticationMethod) (allUsable []config.WorkflowAuthenticationMethod, err error) {
	available := make(map[config.WorkflowAuthenticationMethod]struct{})

	as, err := deps.Authenticators.List(userID)
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		am := a.GetAuthenticationMethod()
		available[am] = struct{}{}
	}

	for _, allowed := range allAllowed {
		_, usable := available[allowed]
		if usable {
			allUsable = append(allUsable, allowed)
		}
	}

	if len(allUsable) == 0 {
		err = NoUsableAuthenticationMethod.NewWithInfo("no usable authentication method", apierrors.Details{
			"allowed": allAllowed,
		})
		return
	}

	return
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
