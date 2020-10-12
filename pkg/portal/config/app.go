package config

import (
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type AppConfig struct {
	HostTemplate string              `envconfig:"HOST_TEMPLATE"`
	IDPattern    string              `envconfig:"ID_PATTERN" default:"^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$"`
	Secret       AppSecretConfig     `envconfig:"SECRET"`
	Kubernetes   AppKubernetesConfig `envconfig:"KUBERNETES"`
	Branding     AppBrandingConfig   `envconfig:"BRANDING"`
}

type TLSCertType string

const (
	TLSCertNone        TLSCertType = "none"
	TLSCertStatic      TLSCertType = "static"
	TLSCertCertManager TLSCertType = "cert-manager"
)

type TLSCertConfig struct {
	Type TLSCertType `envconfig:"TYPE" default:"none"`

	// for static type
	SecretName string `envconfig:"SECRET_NAME"`

	// for cert-manager type
	IssuerKind string `envconfig:"ISSUER_KIND"`
	IssuerName string `envconfig:"ISSUER_NAME"`
}

func (c TLSCertConfig) Validate(ctx *validation.Context) {
	switch c.Type {
	case TLSCertNone:
		return

	case TLSCertStatic:
		if c.SecretName == "" {
			ctx.Child("SECRET_NAME").EmitErrorMessage("missing static TLS secret name")
		}

	case TLSCertCertManager:
		if c.IssuerKind == "" {
			ctx.Child("ISSUER_KIND").EmitErrorMessage("missing cert-manager issuer kind")
		}
		if c.IssuerName == "" {
			ctx.Child("ISSUER_NAME").EmitErrorMessage("missing cert-manager issuer name")
		}

	default:
		if c.SecretName == "" {
			ctx.Child("TYPE").EmitErrorMessage("unknown certificate type")
		}
	}
}

type AppKubernetesConfig struct {
	NewResourcePrefix    string        `envconfig:"NEW_RESOURCE_PREFIX" default:"app-"`
	IngressTemplateFile  string        `envconfig:"INGRESS_TEMPLATE_FILE"`
	DefaultDomainTLSCert TLSCertConfig `envconfig:"DEFAULT_DOMAIN_TLS_CERT"`
	CustomDomainTLSCert  TLSCertConfig `envconfig:"CUSTOM_DOMAIN_TLS_CERT"`
}

type AppSecretConfig struct {
	DatabaseURL      string `envconfig:"DATABASE_URL"`
	DatabaseSchema   string `envconfig:"DATABASE_SCHEMA"`
	RedisURL         string `envconfig:"REDIS_URL"`
	SMTPHost         string `envconfig:"SMTP_HOST"`
	SMTPPort         int    `envconfig:"SMTP_PORT"`
	SMTPMode         string `envconfig:"SMTP_MODE"`
	SMTPUsername     string `envconfig:"SMTP_USERNAME"`
	SMTPPassword     string `envconfig:"SMTP_PASSWORD"`
	TwilioAccountSID string `envconfig:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `envconfig:"TWILIO_AUTH_TOKEN"`
	NexmoAPIKey      string `envconfig:"NEXMO_API_KEY"`
	NexmoAPISecret   string `envconfig:"NEXMO_API_SECRET"`
}

type AppBrandingConfig struct {
	AppName            string `envconfig:"APP_NAME"`
	EmailDefaultSender string `envconfig:"EMAIL_DEFAULT_SENDER"`
	SMSDefaultSender   string `envconfig:"SMS_DEFAULT_SENDER"`
}
