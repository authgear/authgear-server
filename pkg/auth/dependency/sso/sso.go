package sso

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
	// iOS for device iOS
	iOS
	// Android for device Android
	Android
)

// GetURLParams structs parameters for GetLoginAuthURL
type GetURLParams struct {
	Scope       Scope
	Options     Options
	CallbackURL string
	UXMode      UXMode
	UserID      string
	Action      string
}

// Provider defines SSO interface
type Provider interface {
	GetAuthURL(params GetURLParams) (url string, err error)
}
