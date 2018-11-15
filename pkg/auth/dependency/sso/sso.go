package sso

import "strings"

// Scope parameter allows the application to express the desired scope of the access request.
type Scope []string

// Options parameter allows additional options for getting auth url
type Options map[string]interface{}

// UXMode indicates how the URL is used
type UXMode int

const (
	// Undefined for undefined uxmode
	Undefined UXMode = iota
	// WebRedirect for web url redirect
	WebRedirect
	// WebPopup for web popup window
	WebPopup
	// IOS for device iOS
	IOS
	// Android for device Android
	Android
)

func (m UXMode) String() string {
	names := [...]string{
		"undefined",
		"web_redirect",
		"web_popup",
		"ios",
		"android",
	}

	if m < Undefined || m > Android {
		return "undefined"
	}

	return names[m]
}

// GetURLParams structs parameters for GetLoginAuthURL
type GetURLParams struct {
	Scope       Scope
	Options     Options
	CallbackURL string
	UXMode      UXMode
	UserID      string
	Action      string
}

// Config is the base config of a SSO provider
type Config struct {
	Name         string
	Enabled      bool
	ClientID     string
	ClientSecret string
	Scope        Scope
}

// Provider defines SSO interface
type Provider interface {
	GetAuthURL(params GetURLParams) (url string, err error)
}

// NewProvider is the provider factory
func NewProvider(
	name string,
	enabled bool,
	clientID string,
	clientSecret string,
	scopeStr string,
) Provider {
	config := Config{
		Name:         name,
		Enabled:      enabled,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        strings.Split(scopeStr, ","),
	}
	switch name {
	case "google":
		return &GoogleImpl{
			Config: config,
		}
	case "facebook":
		return &FacebookImpl{
			Config: config,
		}
	case "instagram":
		return &InstagramImpl{
			Config: config,
		}
	case "linkedin":
		return &LinkedInImpl{
			Config: config,
		}
	}
	return nil
}
