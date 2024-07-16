package identity

import (
	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

var Deprecated_ErrDuplicatedIdentity = api.NewInvariantViolated("DuplicatedIdentity", "identity already exists", nil)

func IsErrDuplicatedIdentity(err error) bool {
	apiError := apierrors.AsAPIError(err)
	if apiError.Reason == "InvariantViolated" && apiError.HasCause("DuplicatedIdentity") {
		return true
	}
	return false
}

func NewErrDuplicatedIdentity(incoming *Spec, existing *Spec) error {
	details := errorutil.Details{}
	err := api.NewInvariantViolated("DuplicatedIdentity", "identity already exists", nil)

	if incoming != nil {
		details["IdentityTypeIncoming"] = apierrors.APIErrorDetail.Value(incoming.Type)
		switch incoming.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeIncoming"] = apierrors.APIErrorDetail.Value(incoming.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeIncoming"] = apierrors.APIErrorDetail.Value(incoming.OAuth.ProviderID.Type)
		}
	}

	if existing != nil {
		details["IdentityTypeExisting"] = apierrors.APIErrorDetail.Value(existing.Type)
		switch existing.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeExisting"] = apierrors.APIErrorDetail.Value(existing.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeExisting"] = apierrors.APIErrorDetail.Value(existing.OAuth.ProviderID.Type)
		}
	}

	return errorutil.WithDetails(err, details)
}

func NewErrDuplicatedIdentityMany(incoming *Spec, existings []*Spec) error {
	details := errorutil.Details{}
	err := api.NewInvariantViolated("DuplicatedIdentity", "identity already exists", nil)

	if incoming != nil {
		details["IdentityTypeIncoming"] = apierrors.APIErrorDetail.Value(incoming.Type)
		switch incoming.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeIncoming"] = apierrors.APIErrorDetail.Value(incoming.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeIncoming"] = apierrors.APIErrorDetail.Value(incoming.OAuth.ProviderID.Type)
		}
	}

	if len(existings) > 0 {
		// Fill IdentityTypeExisting, LoginIDTypeExisting, OAuthProviderTypeExisting for backward compatibility
		// Use first spec to fill the fields
		firstExistingSpec := existings[0]
		details["IdentityTypeExisting"] = apierrors.APIErrorDetail.Value(firstExistingSpec.Type)
		switch firstExistingSpec.Type {
		case model.IdentityTypeLoginID:
			details["LoginIDTypeExisting"] = apierrors.APIErrorDetail.Value(firstExistingSpec.LoginID.Type)
		case model.IdentityTypeOAuth:
			details["OAuthProviderTypeExisting"] = apierrors.APIErrorDetail.Value(firstExistingSpec.OAuth.ProviderID.Type)
		}

		specDetails := []map[string]interface{}{}
		for _, existingSpec := range existings {
			existingSpec := existingSpec
			thisDetail := map[string]interface{}{}
			thisDetail["IdentityType"] = existingSpec.Type
			switch existingSpec.Type {
			case model.IdentityTypeLoginID:
				thisDetail["LoginIDType"] = existingSpec.LoginID.Type
			case model.IdentityTypeOAuth:
				thisDetail["OAuthProviderType"] = existingSpec.OAuth.ProviderID.Type
			}
			specDetails = append(specDetails, thisDetail)
		}
		details["ExistingIdentities"] = apierrors.APIErrorDetail.Value(specDetails)
	}

	return errorutil.WithDetails(err, details)
}
