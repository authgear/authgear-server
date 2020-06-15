package config

type WelcomeMessageDestination string

const (
	WelcomeMessageDestinationFirst WelcomeMessageDestination = "first"
	WelcomeMessageDestinationAll   WelcomeMessageDestination = "all"
)

type WelcomeMessageConfig struct {
	Enabled      bool                      `json:"enabled,omitempty"`
	EmailMessage EmailMessageConfig        `json:"email_message,omitempty"`
	Destination  WelcomeMessageDestination `json:"destination,omitempty"`
}
