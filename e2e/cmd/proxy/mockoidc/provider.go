package mockoidc

type Provider struct {
	// Authgear's type field for the provider
	Type string

	// OpenID Connect endpoints
	Issuer                string
	AuthorizationEndpoint string
	TokenEndpoint         string
	UserinfoEndpoint      string
	JWKSEndpoint          string
	DiscoveryEndpoint     string

	// Supported values
	ScopesSupported []string
}

var ProviderGoogle = Provider{
	Type:              "google",
	Issuer:            "https://accounts.google.com",
	DiscoveryEndpoint: "https://accounts.google.com/.well-known/openid-configuration",
	ScopesSupported:   []string{"openid", "profile", "email"},
}

var ProviderFacebook = Provider{
	Type:                  "facebook",
	Issuer:                "https://www.facebook.com",
	AuthorizationEndpoint: "https://www.facebook.com/v11.0/dialog/oauth",
	TokenEndpoint:         "https://graph.facebook.com/v11.0/oauth/access_token",
	UserinfoEndpoint:      "https://graph.facebook.com/v11.0/me",
	ScopesSupported:       []string{"public_profile", "email"},
}

var ProviderGithub = Provider{
	Type:                  "github",
	Issuer:                "https://github.com",
	AuthorizationEndpoint: "https://github.com/login/oauth/authorize",
	TokenEndpoint:         "https://github.com/login/oauth/access_token",
	UserinfoEndpoint:      "https://api.github.com/user",
	ScopesSupported:       []string{"read:user", "user:email"},
}

var ProviderLinkedIn = Provider{
	Type:                  "linkedin",
	Issuer:                "https://www.linkedin.com",
	AuthorizationEndpoint: "https://www.linkedin.com/oauth/v2/authorization",
	TokenEndpoint:         "https://www.linkedin.com/oauth/v2/accessToken",
	UserinfoEndpoint:      "https://api.linkedin.com/v2/me",
	ScopesSupported:       []string{"r_liteprofile", "r_emailaddress"},
}

var ProviderADFS = Provider{
	Type:                  "adfs",
	Issuer:                "https://adfs.example.com",
	DiscoveryEndpoint:     "https://adfs.example.com/.well-known/openid-configuration",
	AuthorizationEndpoint: "https://adfs.example.com/oauth2/authorize",
	TokenEndpoint:         "https://adfs.example.com/oauth2/token",
	UserinfoEndpoint:      "https://adfs.example.com/oauth2/userinfo",
	ScopesSupported:       []string{"openid", "profile", "email"},
}

var SupportedProviders = []Provider{
	ProviderGoogle,
	ProviderFacebook,
	ProviderGithub,
	ProviderLinkedIn,
	ProviderADFS,
}
