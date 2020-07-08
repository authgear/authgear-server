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
{{ if .x_password_policies }}
<ul>
{{ range .x_password_policies }}
  {{ if eq .kind "PasswordTooShort" }}
  <li class="primary-txt password-policy length {{ template "PASSWORD_POLICY_CLASS" . }}" data-min-length="{{ .min_length}}">
    {{ localize "password-policy-minimum-length" .min_length }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordUppercaseRequired" }}
  <li class="primary-txt password-policy uppercase {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-uppercase" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordLowercaseRequired" }}
  <li class="primary-txt password-policy lowercase {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-lowercase" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordDigitRequired" }}
  <li class="primary-txt password-policy digit {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-digit" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordSymbolRequired" }}
  <li class="primary-txt password-policy symbol {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-symbol" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordContainingExcludedKeywords" }}
  <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-banned-words" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordBelowGuessableLevel" }}
    {{ if eq .min_level 1.0 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-1" }}
    </li>
    {{ end }}
    {{ if eq .min_level 2.0 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-2" }}
    </li>
    {{ end }}
    {{ if eq .min_level 3.0 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-3" }}
    </li>
    {{ end }}
    {{ if eq .min_level 4.0 }}
    <li class="primary-txt password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-4" }}
    </li>
    {{ end }}
    {{ if eq .min_level 5.0 }}
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
{{- if .x_error_is_password_policy_violated -}}
{{- if .x_is_violated -}}
violated
{{- else -}}
passed
{{- end -}}
{{- end -}}
{{- end -}}
`

const defineError = `
{{ define "ERROR" }}
{{ if .x_error }}
<ul class="errors">
	{{ if eq .x_error.reason "ValidationFailed" }}
		{{ range .x_error.info.causes }}
		{{ if (eq .kind "required") }}
			{{ if (call $.SliceContains .details.missing "x_login_id" ) }}
			<li class="error-txt">{{ localize "error-login-id-required" $.x_login_page_text_login_id_variant }}</li>
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
	{{ else if eq .x_error.reason "InvalidCredentials" }}
		<li class="error-txt">{{ localize "error-invalid-credentials" }}</li>
	{{ else if eq .x_error.reason "PasswordPolicyViolated" }}
		<!-- This error is handled differently -->
	{{ else if eq .x_error.reason "PasswordResetFailed" }}
		<li class="error-txt">{{ localize "error-password-reset-failed" }}</li>
	{{ else if eq .x_error.reason "DuplicatedIdentity" }}
		<li class="error-txt">{{ localize "error-duplicated-identity" }}</li>
	{{ else if eq .x_error.reason "InvalidIdentityRequest" }}
		<li class="error-txt">{{ localize "error-remove-last-identity" }}</li>
	{{ else }}
		<li class="error-txt">{{ .x_error.message }}</li>
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
	Default: `
{{ define "auth_ui_html_head.html" }}
<head>
<title>{{ .app_name }}</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="{{ .x_static_asset_url_prefix }}/authui/css/main.css">
<script src="{{ .x_static_asset_url_prefix }}/authui/js/main.js"></script>
{{ if .x_css }}
<style>
{{ .x_css }}
</style>
{{ end }}
</head>
{{ end }}
`,
}

var TemplateAuthUIHeaderHTML = template.Spec{
	Type:   TemplateItemTypeAuthUIHeaderHTML,
	IsHTML: true,
	Default: `
{{ define "auth_ui_header.html" }}
{{ if .logo_uri }}
<div class="logo" style="background-image: url('{{ .logo_uri }}'); background-position: center; background-size: contain; background-repeat: no-repeat"></div>
{{ else }}
<div class="logo"></div>
{{ end }}
{{ end }}
`,
}

var TemplateAuthUIFooterHTML = template.Spec{
	Type:   TemplateItemTypeAuthUIFooterHTML,
	IsHTML: true,
	Default: `
{{ define "auth_ui_footer.html" }}
<div class="authgear-logo"></div>
{{ end }}
`,
}

var components = []config.TemplateItemType{
	TemplateItemTypeAuthUIHTMLHeadHTML,
	TemplateItemTypeAuthUIHeaderHTML,
	TemplateItemTypeAuthUIFooterHTML,
}
