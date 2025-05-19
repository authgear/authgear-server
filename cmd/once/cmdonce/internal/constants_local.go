//go:build authgearonce_license_server_local

package internal

const (
	LicenseServerEndpoint            = "http://localhost:8200"
	LicenseServerEndpointOverridable = true

	QuestionName_EnableCertbot_PromptEnabled            = false
	QuestionName_SelectCertbotEnvironment_PromptEnabled = false

	KeepInstallationContainerByDefault = true
)
