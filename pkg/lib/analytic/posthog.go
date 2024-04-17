package analytic

type PosthogCredentials struct {
	Endpoint string
	APIKey   string
}

type PosthogIntegration struct {
	PosthogCredentials *PosthogCredentials
}

func (p *PosthogIntegration) SetGroupProperties() error {
	// TODO
	return nil
}
