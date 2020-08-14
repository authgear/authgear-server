package authenticator

const (
	// TagPrimaryAuthenticator indicates the authenticator is primary.
	TagPrimaryAuthenticator string = "authentication:primary_authenticator"
	// TagSecondaryAuthenticator indicates the authenticator is secondary.
	TagSecondaryAuthenticator string = "authentication:secondary_authenticator"
	// TagDefaultAuthenticator indicates the authenticator is the default one to use in MFA.
	TagDefaultAuthenticator string = "authentication:default_authenticator"
)
