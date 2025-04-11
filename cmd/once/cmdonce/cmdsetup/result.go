package cmdsetup

type SetupAppResult struct {
	AUTHGEAR_HTTP_ORIGIN_PROJECT      string
	AUTHGEAR_HTTP_ORIGIN_PORTAL       string
	AUTHGEAR_HTTP_ORIGIN_ACCOUNTS     string
	AUTHGEAR_ONCE_ADMIN_USER_EMAIL    string
	AUTHGEAR_ONCE_ADMIN_USER_PASSWORD string
	// "production" or "staging"
	AUTHGEAR_CERTBOT_ENVIRONMENT string

	CertbotEnabled bool

	SMTPHost          string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	SMTPSenderAddress string
}
