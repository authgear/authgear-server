//go:build authgearonce_license_server_staging

package internal

const (
	LicenseServerEndpoint            = "https://once-license-staging.authgear.com"
	LicenseServerEndpointOverridable = true

	QuestionName_EnableCertbot_PromptEnabled            = false
	QuestionName_SelectCertbotEnvironment_PromptEnabled = false

	KeepInstallationContainerByDefault = false
)
