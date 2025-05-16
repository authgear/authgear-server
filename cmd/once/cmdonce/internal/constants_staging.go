//go:build authgearonce_license_server_staging

package internal

const (
	LicenseServerEndpoint            = "https://once-license-staging.authgear.com"
	LicenseServerEndpointOverridable = true

	QuestionName_EnableCertbot_PromptByDefault            = true
	QuestionName_SelectCertbotEnvironment_PromptByDefault = false
)
