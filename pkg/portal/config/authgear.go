package config

type AuthgearConfig struct {
	ClientID string `envconfig:"CLIENT_ID"`

	// Endpoint is the endpoint for public connection.
	// Its primary usage is to be included in system-config.json and consumed by the portal.
	Endpoint string `envconfig:"ENDPOINT"`
	// EndpointInternal is an environment variable acting as an alternative to Endpoint.
	// In Authgear ONCE, Endpoint resolves to a public IP address, which may not be accessible within the container.
	// When EndpointInternal is present, it is used instead.
	EndpointInternal string `envconfig:"ENDPOINT_INTERNAL"`

	AppID string `envconfig:"APP_ID"`

	WebSDKSessionType string `envconfig:"WEB_SDK_SESSION_TYPE" default:"cookie"`
}
