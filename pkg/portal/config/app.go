package config

import (
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type AppConfig struct {
	HostSuffix string              `envconfig:"HOST_SUFFIX"`
	IDPattern  string              `envconfig:"ID_PATTERN" default:"^[a-z0-9][a-z0-9-]{2,30}[a-z0-9]$"`
	Kubernetes AppKubernetesConfig `envconfig:"KUBERNETES"`

	// BuiltinResourceDirectory sets the directory for built-in resource files
	BuiltinResourceDirectory string `envconfig:"BUILTIN_RESOURCE_DIRECTORY" default:"resources/authgear"`
	// CustomResourceDirectory sets the directory for customized resource files
	CustomResourceDirectory string `envconfig:"CUSTOM_RESOURCE_DIRECTORY"`
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
	IngressTemplateFile  string        `envconfig:"INGRESS_TEMPLATE_FILE"`
	DefaultDomainTLSCert TLSCertConfig `envconfig:"DEFAULT_DOMAIN_TLS_CERT"`
	CustomDomainTLSCert  TLSCertConfig `envconfig:"CUSTOM_DOMAIN_TLS_CERT"`
}
