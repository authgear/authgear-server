package webapp

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypeAuthUIHTMLHeadHTML       string = "auth_ui_html_head.html"
	TemplateItemTypeAuthUIHeaderHTML         string = "auth_ui_header.html"
	TemplateItemTypeAuthUINavBarHTML         string = "auth_ui_nav_bar.html"
	TemplateItemTypeAuthUIPasswordPolicyHTML string = "auth_ui_password_policy.html"
)

// nolint: gosec
const definePasswordPolicyClass = `
{{- define "PASSWORD_POLICY_CLASS" -}}
{{- if .Info.x_error_is_password_policy_violated -}}
{{- if .Info.x_is_violated -}}
error-txt
{{- else -}}
good-txt
{{- end -}}
{{- end -}}
{{- end -}}
`

const defineError = `
{{ define "ERROR" }}
{{ if .Error }}
<ul class="errors">
	{{ if eq .Error.reason "ValidationFailed" }}
		{{ range .Error.info.causes }}
		{{ if (eq .kind "required") }}
			{{ if (call $.SliceContains .details.missing "x_login_id" ) }}
			<li class="error-txt">{{ template "error-login-id-required" (makemap "variant" $.LoginPageTextLoginIDVariant) }}</li>
			{{ else if (call $.SliceContains .details.missing "x_password" ) }}
			<li class="error-txt">{{ template "error-password-or-code-required" }}</li>
			{{ else if (call $.SliceContains .details.missing "x_calling_code" ) }}
			<li class="error-txt">{{ template "error-calling-code-required" }}</li>
			{{ else if (call $.SliceContains .details.missing "x_national_number" ) }}
			<li class="error-txt">{{ template "error-phone-number-required" }}</li>
			{{ else }}
			<li class="error-txt">{{ . }}</li>
			{{ end }}
		{{ else if (eq .kind "format") }}
			{{ if (eq .details.format "phone") }}
			<li class="error-txt">{{ template "error-phone-number-format" }}</li>
			{{ else if (eq .details.format "email") }}
			<li class="error-txt">{{ template "error-invalid-email" }}</li>
			{{ else if (eq .details.format "username") }}
			<li class="error-txt">{{ template "error-invalid-username" }}</li>
			{{ else }}
			<li class="error-txt">{{ . }}</li>
			{{ end }}
		{{ else }}
		<li class="error-txt">{{ . }}</li>
		{{ end }}
		{{ end }}
	{{ else if eq .Error.reason "InvalidCredentials" }}
		<li class="error-txt">{{ template "error-invalid-credentials" }}</li>
	{{ else if eq .Error.reason "PasswordPolicyViolated" }}
		<!-- This error is handled differently -->
	{{ else if eq .Error.reason "PasswordResetFailed" }}
		<li class="error-txt">{{ template "error-password-reset-failed" }}</li>
	{{ else if eq .Error.reason "DuplicatedIdentity" }}
		<li class="error-txt">{{ template "error-duplicated-identity" }}</li>
	{{ else if eq .Error.reason "InvalidIdentityRequest" }}
		<li class="error-txt">{{ template "error-remove-last-identity" }}</li>
	{{ else }}
		<li class="error-txt">{{ .Error.message }}</li>
	{{ end }}
</ul>
{{ end }}
{{ end }}
`

var defines = []string{
	defineError,
	definePasswordPolicyClass,
}

var TemplateAuthUIHTMLHeadHTML = template.Register(template.T{
	Type:   TemplateItemTypeAuthUIHTMLHeadHTML,
	IsHTML: true,
})

var TemplateAuthUIHeaderHTML = template.Register(template.T{
	Type:   TemplateItemTypeAuthUIHeaderHTML,
	IsHTML: true,
})

var TemplateAuthUINavBarHTML = template.Register(template.T{
	Type:   TemplateItemTypeAuthUINavBarHTML,
	IsHTML: true,
})

var TemplateAuthUIPasswordPolicyHTML = template.Register(template.T{
	Type:   TemplateItemTypeAuthUIPasswordPolicyHTML,
	IsHTML: true,
})

var components = []string{
	TemplateItemTypeAuthUIHTMLHeadHTML,
	TemplateItemTypeAuthUIHeaderHTML,
	TemplateItemTypeAuthUINavBarHTML,
	TemplateItemTypeAuthUIPasswordPolicyHTML,
}
