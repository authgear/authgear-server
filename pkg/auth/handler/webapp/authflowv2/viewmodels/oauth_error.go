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
		} else if rawError.Reason == "InvariantViolated" {
			cause, ok := rawError.Info["cause"].(map[string]interface{})
			if !ok {
				return nil
			}
			kind, ok := cause["kind"].(string)
			if !ok {
				return nil
			}
			if kind == "DuplicatedIdentity" {
				return rawError
			}
		}

		return nil
	}

	return OAuthErrorViewModel{
		OAuthError: getOAuthError(),
	}
}
