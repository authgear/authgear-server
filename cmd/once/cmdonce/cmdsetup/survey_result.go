package cmdsetup

type SurveyResult struct {
	DomainProject  string
	DomainPortal   string
	DomainAccounts string

	CertbotEnabled bool
	// "production" or "staging"
	CertbotEnvironment string

	AdminEmailAddress string
	AdminPassword     string

	SMTPHost          string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	SMTPSenderAddress string
}
