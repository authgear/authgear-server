package sso

// UXMode affects how response is delivered.
type UXMode string

const (
	// UXModeWebRedirect delivers the response via HTTP 302.
	UXModeWebRedirect UXMode = "web_redirect"
	// UXModeWebPopup delivers the response by rendering an HTML with some JavaScript to use window.postMessage.
	UXModeWebPopup UXMode = "web_popup"
	// UXModeMobileApp delivers the response via HTTP 302.
	UXModeMobileApp UXMode = "mobile_app"
	// UXModeManual delivers the response directly in the HTTP response body in JSON.
	UXModeManual UXMode = "manual"
)

// IsValidUXMode validates UXMode
func IsValidUXMode(mode UXMode) bool {
	allModes := []UXMode{UXModeWebRedirect, UXModeWebPopup, UXModeMobileApp, UXModeManual}
	for _, v := range allModes {
		if mode == v {
			return true
		}
	}
	return false
}
