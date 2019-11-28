package sso

// UXMode indicates how the URL is used
type UXMode string

// UXMode constants
const (
	UXModeWebRedirect UXMode = "web_redirect"
	UXModeWebPopup    UXMode = "web_popup"
	UXModeMobileApp   UXMode = "mobile_app"
)

// IsValidUXMode validates UXMode
func IsValidUXMode(mode UXMode) bool {
	allModes := []UXMode{UXModeWebRedirect, UXModeWebPopup, UXModeMobileApp}
	for _, v := range allModes {
		if mode == v {
			return true
		}
	}
	return false
}
