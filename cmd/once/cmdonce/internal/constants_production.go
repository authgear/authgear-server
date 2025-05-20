//go:build !(authgearonce_license_server_local || authgearonce_license_server_staging)

package internal

const (
	LicenseServerEndpoint            = "https://once-license.authgear.com"
	LicenseServerEndpointOverridable = false

	QuestionName_EnableCertbot_PromptEnabled            = false
	QuestionName_SelectCertbotEnvironment_PromptEnabled = false

	KeepInstallationContainerByDefault = false
)
