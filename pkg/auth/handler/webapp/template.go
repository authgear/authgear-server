package webapp

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	TemplateItemTypeAuthUIHTMLHeadHTML config.TemplateItemType = "auth_ui_html_head.html"
	TemplateItemTypeAuthUIHeaderHTML   config.TemplateItemType = "auth_ui_header.html"
	TemplateItemTypeAuthUIFooterHTML   config.TemplateItemType = "auth_ui_footer.html"
)

// nolint: gosec
const definePasswordPolicy = `
{{ define "PASSWORD_POLICY" }}
{{ if .PasswordPolicies }}
<ul>
{{ range .PasswordPolicies }}
  {{ if eq .Name "PasswordTooShort" }}
  <li class="primary-txt password-policy length {{ template "PASSWORD_POLICY_CLASS" . }}" data-min-length="{{ .Info.min_length}}">
    {{ localize "password-policy-minimum-length" .Info.min_length }}
  </li>
  {{ end }}
  {{ if eq .Name "PasswordUppercaseRequired" }}
  <li class="primary-txt password-policy uppercase {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-uppercase" }}
  </li>
  {{ end }}
  {{ if eq .Name "PasswordLowercaseRequired" }}
  <li class="primary-txt password-policy lowercase {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-lowercase" }}
  </li>
  {{ end }}
  {{ if eq .Name "PasswordDigitRequired" }}
  <li class="primary-txt password-policy digit {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-digit" }}
  </li>
  {{ end }}
  {{ if eq .Name "PasswordSymbolRequired" }}
  <li class="primary-txt password-policy symbol {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-symbol" }}
  </li>
  {{ end }}
  {{ if eq .Name "PasswordContainingExcludedKeywords" }}
  <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-banned-words" }}
  </li>
  {{ end }}
  {{ if eq .Name "PasswordReused" }}
  <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-reuse" .Info.history_size .Info.history_days }}
  </li>
  {{ end }}
  {{ if eq .Name "PasswordBelowGuessableLevel" }}
    {{ if eq .Info.min_level 1 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-1" }}
    </li>
    {{ end }}
    {{ if eq .Info.min_level 2 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-2" }}
    </li>
    {{ end }}
    {{ if eq .Info.min_level 3 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-3" }}
    </li>
    {{ end }}
    {{ if eq .Info.min_level 4 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-4" }}
    </li>
    {{ end }}
    {{ if eq .Info.min_level 5 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-5" }}
    </li>
    {{ end }}
  {{ end }}
{{ end }}
</ul>
{{ end }}
{{ end }}
`

// nolint: gosec
const definePasswordPolicyClass = `
{{- define "PASSWORD_POLICY_CLASS" -}}
{{- if .Info.x_error_is_password_policy_violated -}}
{{- if .Info.x_is_violated -}}
violated
{{- else -}}
passed
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
			<li class="error-txt">{{ localize "error-login-id-required" $.LoginPageTextLoginIDVariant }}</li>
			{{ else if (call $.SliceContains .details.missing "x_password" ) }}
			<li class="error-txt">{{ localize "error-password-or-code-required" }}</li>
			{{ else if (call $.SliceContains .details.missing "x_calling_code" ) }}
			<li class="error-txt">{{ localize "error-calling-code-required" }}</li>
			{{ else if (call $.SliceContains .details.missing "x_national_number" ) }}
			<li class="error-txt">{{ localize "error-phone-number-required" }}</li>
			{{ else }}
			<li class="error-txt">{{ . }}</li>
			{{ end }}
		{{ else if (eq .kind "format") }}
			{{ if (eq .details.format "phone") }}
			<li class="error-txt">{{ localize "error-phone-number-format" }}</li>
			{{ else if (eq .details.format "email") }}
			<li class="error-txt">{{ localize "error-invalid-email" }}</li>
			{{ else if (eq .details.format "username") }}
			<li class="error-txt">{{ localize "error-invalid-username" }}</li>
			{{ else }}
			<li class="error-txt">{{ . }}</li>
			{{ end }}
		{{ else }}
		<li class="error-txt">{{ . }}</li>
		{{ end }}
		{{ end }}
	{{ else if eq .Error.reason "InvalidCredentials" }}
		<li class="error-txt">{{ localize "error-invalid-credentials" }}</li>
	{{ else if eq .Error.reason "PasswordPolicyViolated" }}
		<!-- This error is handled differently -->
	{{ else if eq .Error.reason "PasswordResetFailed" }}
		<li class="error-txt">{{ localize "error-password-reset-failed" }}</li>
	{{ else if eq .Error.reason "DuplicatedIdentity" }}
		<li class="error-txt">{{ localize "error-duplicated-identity" }}</li>
	{{ else if eq .Error.reason "InvalidIdentityRequest" }}
		<li class="error-txt">{{ localize "error-remove-last-identity" }}</li>
	{{ else }}
		<li class="error-txt">{{ .Error.message }}</li>
	{{ end }}
</ul>
{{ end }}
{{ end }}
`

var defines = []string{
	defineError,
	definePasswordPolicy,
	definePasswordPolicyClass,
}

var TemplateAuthUIHTMLHeadHTML = template.Spec{
	Type:   TemplateItemTypeAuthUIHTMLHeadHTML,
	IsHTML: true,
}

var TemplateAuthUIHeaderHTML = template.Spec{
	Type:   TemplateItemTypeAuthUIHeaderHTML,
	IsHTML: true,
}

var TemplateAuthUIFooterHTML = template.Spec{
	Type:   TemplateItemTypeAuthUIFooterHTML,
	IsHTML: true,
}

var components = []config.TemplateItemType{
	TemplateItemTypeAuthUIHTMLHeadHTML,
	TemplateItemTypeAuthUIHeaderHTML,
	TemplateItemTypeAuthUIFooterHTML,
}
