//go:build authgearonce_license_server_local

package internal

const (
	LicenseServerEndpoint            = "http://localhost:8200"
	LicenseServerEndpointOverridable = true

	QuestionName_EnableCertbot_PromptByDefault            = true
	QuestionName_SelectCertbotEnvironment_PromptByDefault = true

	KeepInstallationContainerByDefault = true
)
