package sso

type UserProfile struct {
	ProviderRawProfile map[string]interface{}
	// ProviderUserID is not necessarily equal to sub.
	// If there exists a more unique identifier than sub, that identifier is chosen instead.
	ProviderUserID     string
	StandardAttributes map[string]interface{}
}
