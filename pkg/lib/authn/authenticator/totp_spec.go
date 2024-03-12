package authenticator

type TOTPSpec struct {
	Code        string `json:"code,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Secret      string `json:"-"`
}
