package viewmodels

import "github.com/authgear/authgear-server/pkg/api/apierrors"

type OAuthErrorViewModel struct {
	OAuthError error
}

func NewOAuthErrorViewModel(rawError *apierrors.APIError) OAuthErrorViewModel {

	getOAuthError := func() error {
		if rawError == nil {
			return nil
		}
		if rawError.Reason == "UserNotFound" && rawError.Info["IdentityTypeExisting"] != nil {
			return rawError
		} else if rawError.Reason == "DuplicatedIdentity" {
			return rawError
		}

		return nil
	}

	return OAuthErrorViewModel{
		OAuthError: getOAuthError(),
	}
}
