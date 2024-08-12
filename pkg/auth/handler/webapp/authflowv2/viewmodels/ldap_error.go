package viewmodels

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/handler/webapp/viewmodels"
)

type LDAPInputErrorViewModel struct {
	LDAPUsernameInputError *InputError
	PasswordInputError     *InputError
	HasUnknownError        bool
}

// nolint: gocognit
func NewLDAPInputErrorViewModel(apiError *apierrors.APIError) LDAPInputErrorViewModel {
	viewModel := LDAPInputErrorViewModel{
		LDAPUsernameInputError: &InputError{
			HasError:        false,
			HasErrorMessage: false,
		},
		PasswordInputError: &InputError{
			HasError:        false,
			HasErrorMessage: false,
		},
	}
	if apiError != nil {
		switch apiError.Reason {
		case "InvalidCredentials":
			viewModel.LDAPUsernameInputError.HasError = true
			viewModel.PasswordInputError.HasError = true
			// Alert invalid credentials error
			viewModel.HasUnknownError = true
		case "ValidationFailed":
			for _, causes := range apiError.Info["causes"].([]interface{}) {
				if cause, ok := causes.(map[string]interface{}); ok {
					if kind, ok := cause["kind"].(string); ok {
						if kind == "required" {
							if details, ok := cause["details"].(map[string]interface{}); ok {
								if missing, ok := details["missing"].([]interface{}); ok {
									if viewmodels.SliceContains(missing, "x_username") {
										viewModel.LDAPUsernameInputError.HasError = true
										viewModel.LDAPUsernameInputError.HasErrorMessage = true
									} else if viewmodels.SliceContains(missing, "x_password") {
										viewModel.PasswordInputError.HasError = true
										viewModel.PasswordInputError.HasErrorMessage = true
									}
								}
							}
						}
					}
				}
			}
		}

		if !viewModel.LDAPUsernameInputError.HasError && !viewModel.PasswordInputError.HasError {
			// If it is not an error shown in inputs, it is an unknown error
			viewModel.HasUnknownError = true
		}
	}

	return viewModel
}
