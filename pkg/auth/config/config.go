package config

type AppConfig struct {
	ID       string      `json:"id"`
	Metadata AppMetadata `json:"metadata,omitempty"`

	HTTP *HTTPConfig `json:"http,omitempty"`
	Hook *HookConfig `json:"hook,omitempty"`

	Template *TemplateConfig `json:"template,omitempty"`
	UI       *UIConfig       `json:"ui,omitempty"`

	Authentication *AuthenticationConfig `json:"authentication,omitempty"`
	Session        *SessionConfig        `json:"session,omitempty"`
	OAuth          *OAuthConfig          `json:"oauth,omitempty"`
	Identity       *IdentityConfig       `json:"identity,omitempty"`
	Authenticator  *AuthenticatorConfig  `json:"authenticator,omitempty"`

	ForgotPassword *ForgotPasswordConfig `json:"forgot_password,omitempty"`
	WelcomeMessage *WelcomeMessageConfig `json:"welcome_message,omitempty"`
}
