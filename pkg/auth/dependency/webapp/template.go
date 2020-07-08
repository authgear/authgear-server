package webapp

import (
	"github.com/authgear/authgear-server/pkg/auth/config"
	"github.com/authgear/authgear-server/pkg/template"
)

const (
	// Components
	TemplateItemTypeAuthUIHTMLHeadHTML config.TemplateItemType = "auth_ui_html_head.html"
	TemplateItemTypeAuthUIHeaderHTML   config.TemplateItemType = "auth_ui_header.html"
	TemplateItemTypeAuthUIFooterHTML   config.TemplateItemType = "auth_ui_footer.html"

	// Interaction entrypoints
	TemplateItemTypeAuthUILoginHTML   config.TemplateItemType = "auth_ui_login.html"
	TemplateItemTypeAuthUISignupHTML  config.TemplateItemType = "auth_ui_signup.html"
	TemplateItemTypeAuthUIPromoteHTML config.TemplateItemType = "auth_ui_promote.html"

	// Interaction steps
	// nolint: gosec
	TemplateItemTypeAuthUIEnterPasswordHTML config.TemplateItemType = "auth_ui_enter_password.html"
	// nolint: gosec
	TemplateItemTypeAuthUICreatePasswordHTML config.TemplateItemType = "auth_ui_create_password.html"
	TemplateItemTypeAuthUIOOBOTPHTML         config.TemplateItemType = "auth_ui_oob_otp_html"
	TemplateItemTypeAuthUIEnterLoginIDHTML   config.TemplateItemType = "auth_ui_enter_login_id.html"

	// Forgot Password
	// nolint: gosec
	TemplateItemTypeAuthUIForgotPasswordHTML config.TemplateItemType = "auth_ui_forgot_password.html"
	// nolint: gosec
	TemplateItemTypeAuthUIForgotPasswordSuccessHTML config.TemplateItemType = "auth_ui_forgot_password_success.html"
	// nolint: gosec
	TemplateItemTypeAuthUIResetPasswordHTML config.TemplateItemType = "auth_ui_reset_password.html"
	// nolint: gosec
	TemplateItemTypeAuthUIResetPasswordSuccessHTML config.TemplateItemType = "auth_ui_reset_password_success.html"

	// Logout
	TemplateItemTypeAuthUILogoutHTML config.TemplateItemType = "auth_ui_logout.html"

	// Settings
	TemplateItemTypeAuthUISettingsHTML         config.TemplateItemType = "auth_ui_settings.html"
	TemplateItemTypeAuthUISettingsIdentityHTML config.TemplateItemType = "auth_ui_settings_identity.html"
)

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

var defines = []string{
	defineError,
	definePasswordPolicy,
	definePasswordPolicyClass,
}

var components = []config.TemplateItemType{
	TemplateItemTypeAuthUIHTMLHeadHTML,
	TemplateItemTypeAuthUIHeaderHTML,
	TemplateItemTypeAuthUIFooterHTML,
}

var TemplateAuthUILoginHTML = template.Spec{
	Type:        TemplateItemTypeAuthUILoginHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
	<div class="content">
		{{ template "auth_ui_header.html" . }}
		<div class="authorize-form">
			<div class="authorize-idp-section">
				{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				<form class="authorize-idp-form" method="post" novalidate>
				{{ $.csrfField }}
				<button class="btn sso-btn {{ .provider_type }}" type="submit" name="x_idp_id" value="{{ .provider_alias }}" data-form-xhr="false">
					{{- if eq .provider_type "apple" -}}
					{{ localize "sign-in-apple" }}
					{{- end -}}
					{{- if eq .provider_type "google" -}}
					{{ localize "sign-in-google" }}
					{{- end -}}
					{{- if eq .provider_type "facebook" -}}
					{{ localize "sign-in-facebook" }}
					{{- end -}}
					{{- if eq .provider_type "linkedin" -}}
					{{ localize "sign-in-linkedin" }}
					{{- end -}}
					{{- if eq .provider_type "azureadv2" -}}
					{{ localize "sign-in-azureadv2" }}
					{{- end -}}
				</button>
				</form>
				{{ end }}
				{{ end }}
			</div>

			{{ $has_oauth := false }}
			{{ $has_login_id := false }}
			{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				{{ $has_oauth = true }}
				{{ end }}
				{{ if eq .type "login_id" }}
				{{ $has_login_id = true }}
				{{ end }}
			{{ end }}
			{{ if $has_oauth }}{{ if $has_login_id }}
			<div class="primary-txt sso-loginid-separator">{{ localize "sso-login-id-separator" }}</div>
			{{ end }}{{ end }}

			{{ template "ERROR" . }}

			<form class="authorize-loginid-form" method="post" novalidate>
				{{ $.csrfField }}

				{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if .x_login_page_login_id_has_phone }}
				<div class="phone-input">
					<select class="input select primary-txt" name="x_calling_code">
						{{ range .x_calling_codes }}
						<option
							value="{{ . }}"
							{{ if $.x_calling_code }}{{ if eq $.x_calling_code . }}
							selected
							{{ end }}{{ end }}
							>
							+{{ . }}
						</option>
						{{ end }}
					</select>
					<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}">
				</div>
				{{ end }}{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if not (eq .x_login_id_input_type "phone") }}{{ if (not (eq .x_login_page_text_login_id_variant "none")) }}
				<input class="input text-input primary-txt" type="{{ .x_login_page_text_login_id_input_type }}" name="x_login_id" placeholder="{{ localize "login-id-placeholder" .x_login_page_text_login_id_variant }}">
				{{ end }}{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if (not (eq .x_login_page_text_login_id_variant "none")) }}
				<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" .x_login_page_text_login_id_input_type }}">{{ localize "use-text-login-id-description" .x_login_page_text_login_id_variant }}</a>
				{{ end }}{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if not (eq .x_login_id_input_type "phone") }}{{ if .x_login_page_login_id_has_phone }}
				<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
				{{ end }}{{ end }}{{ end }}

				<div class="link">
					<span class="primary-text">{{ localize "signup-button-hint" }}</span>
					<a href="{{ call .MakeURLWithPathWithoutX "/signup" }}">{{ localize "signup-button-label" }}</a>
				</div>

				{{ if .x_password_authenticator_enabled }}
				<a class="link align-self-flex-start" href="{{ call .MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>
				{{ end }}

				{{ if or .x_login_page_login_id_has_phone (not (eq .x_login_page_text_login_id_variant "none")) }}
				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
				{{ end }}
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

var TemplateAuthUIEnterPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIEnterPasswordHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<form class="simple-form vertical-form form-fields-container" method="post" novalidate>
{{ $.csrfField }}

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
	<div class="login-id primary-txt">
	{{ if .x_national_number }}
		+{{ .x_calling_code}} {{ .x_national_number }}
	{{ else }}
		{{ .x_login_id }}
	{{ end }}
	</div>
</div>

<div class="title primary-txt">{{ localize "enter-password-page-title" }}</div>

{{ template "ERROR" . }}

<input id="password" class="input text-input primary-txt" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

{{ if .x_password_authenticator_enabled }}
<a class="link align-self-flex-start" href="{{ call .MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label--enter-password-page" }}</a>
{{ end }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUIOOBOTPHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIOOBOTPHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="simple-form vertical-form form-fields-container">

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<!-- FIXME: x_login_id_input_type, x_calling_code, x_national_number, x_login_id -->

{{ if eq .x_login_id_input_type "phone" }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--sms" }}</div>
{{ end }}
{{ if not (eq .x_login_id_input_type "phone") }}
<div class="title primary-txt">{{ localize "oob-otp-page-title--email" }}</div>
{{ end }}

{{ template "ERROR" . }}

{{ if eq .x_login_id_input_type "phone" }}
<!-- FIXME: x_calling_code x_national_number -->
<div class="description primary-txt">{{ localize "oob-otp-description--sms" .x_oob_otp_code_length "FIXME" "FIXME" }}</div>
{{ end }}
{{ if not (eq .x_login_id_input_type "phone") }}
<!-- FIXME: x_login_id -->
<div class="description primary-txt">{{ localize "oob-otp-description--email" .x_oob_otp_code_length "FIXME" }}</div>
{{ end }}

<form class="vertical-form form-fields-container" method="post" novalidate>
{{ $.csrfField }}

<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_password" placeholder="{{ localize "oob-otp-placeholder" }}">
<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
</form>

<form class="link oob-otp-trigger-form" method="post" novalidate>
{{ $.csrfField }}

<span class="primary-txt">{{ localize "oob-otp-resend-button-hint" }}</span>
<button id="resend-button" class="anchor" type="submit" name="trigger" value="true"
	data-cooldown="{{ .x_oob_otp_code_send_cooldown }}"
	data-label="{{ localize "oob-otp-resend-button-label" }}"
	data-label-unit="{{ localize "oob-otp-resend-button-label--unit" }}">{{ localize "oob-otp-resend-button-label" }}</button>
</form>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUIEnterLoginIDHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIEnterLoginIDHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="simple-form vertical-form form-fields-container">

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<!-- FIXME: x_old_login_id_value, x_login_id_key, x_login_id_type, x_login_id_input_type -->

<div class="title primary-txt">
	{{ if .x_old_login_id_value }}
	{{ localize "enter-login-id-page-title--change" .x_login_id_key }}
	{{ else }}
	{{ localize "enter-login-id-page-title--add" .x_login_id_key }}
	{{ end }}
</div>

{{ template "ERROR" . }}

<form class="vertical-form form-fields-container" method="post" novalidate>

{{ $.csrfField }}
<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">
<input type="hidden" name="x_login_id_type" value="{{ .x_login_id_type }}">
<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">
<input type="hidden" name="x_old_login_id_value" value="{{ .x_old_login_id_value }}">

{{ if eq .x_login_id_input_type "phone" }}
<div class="phone-input">
	<select class="input select primary-txt" name="x_calling_code">
		{{ range .x_calling_codes }}
		<option
			value="{{ . }}"
			{{ if $.x_calling_code }}{{ if eq $.x_calling_code . }}
			selected
			{{ end }}{{ end }}
			>
			+{{ . }}
		</option>
		{{ end }}
	</select>
	<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}">
</div>
{{ else }}
<input class="input text-input primary-txt" type="{{ .x_login_id_input_type }}" name="x_login_id" placeholder="{{ localize "login-id-placeholder" .x_login_id_type }}">
{{ end }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>

{{ if .x_old_login_id_value }}
<form class="enter-login-id-remove-form" method="post" novalidate>
{{ $.csrfField }}
<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">
<input type="hidden" name="x_login_id_type" value="{{ .x_login_id_type }}">
<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">
<input type="hidden" name="x_old_login_id_value" value="{{ .x_old_login_id_value }}">
<button class="anchor" type="submit" name="x_action" value="remove">{{ localize "disconnect-button-label" }}</button>
{{ end }}
</form>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUIForgotPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIForgotPasswordHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<form class="simple-form vertical-form form-fields-container" method="post" novalidate>
{{ $.csrfField }}

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<div class="title primary-txt">{{ localize "forgot-password-page-title" }}</div>

{{ template "ERROR" . }}

{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if .x_login_page_login_id_has_phone }}
<div class="description primary-txt">{{ localize "forgot-password-phone-description" }}</div>
<div class="phone-input">
	<select class="input select primary-txt" name="x_calling_code">
		{{ range .x_calling_codes }}
		<option
			value="{{ . }}"
			{{ if $.x_calling_code }}{{ if eq $.x_calling_code . }}
			selected
			{{ end }}{{ end }}
			>
			+{{ . }}
		</option>
		{{ end }}
	</select>
	<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}">
</div>
{{ end }}{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if (not (eq .x_login_id_input_type "phone")) }}{{ if or (eq .x_login_page_text_login_id_variant "email") (eq .x_login_page_text_login_id_variant "email_or_username") }}
<div class="description primary-txt">{{ localize "forgot-password-email-description" }}</div>
<input class="input text-input primary-txt" type="{{ .x_login_id_input_type }}" name="x_login_id" placeholder="{{ localize "email-placeholder" }}">
{{ end }}{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "phone" }}{{ if or (eq .x_login_page_text_login_id_variant "email") (eq .x_login_page_text_login_id_variant "email_or_username") }}
<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "email" }}">{{ localize "use-email-login-id-description" }}</a>
{{ end }}{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if eq .x_login_id_input_type "email" }}{{ if .x_login_page_login_id_has_phone }}
<a class="link align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
{{ end }}{{ end }}{{ end }}

{{ if or .x_login_page_login_id_has_phone (not (eq .x_login_page_text_login_id_variant "none")) }}
<button class="btn primary-btn submit-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>
{{ end }}

</form>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUIForgotPasswordSuccessHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIForgotPasswordSuccessHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="simple-form vertical-form form-fields-container">

<div class="title primary-txt">{{ localize "forgot-password-success-page-title" }}</div>

{{ template "ERROR" . }}

<!-- FIXME: x_login_id -->

<div class="description primary-txt">{{ localize "forgot-password-success-description" .x_login_id }}</div>

<a class="btn primary-btn align-self-flex-end" href="{{ call .MakeURLWithPathWithoutX "/login" }}">{{ localize "login-button-label--forgot-password-success-page" }}</a>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUIResetPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIResetPasswordHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<form class="simple-form vertical-form form-fields-container" method="post" novalidate>
{{ $.csrfField }}

<div class="title primary-txt">{{ localize "reset-password-page-title" }}</div>

{{ template "ERROR" . }}

<div class="description primary-txt">{{ localize "reset-password-description" }}</div>

<input id="password" data-password-policy-password="" class="input text-input primary-txt" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

{{ template "PASSWORD_POLICY" . }}

<button class="btn primary-btn submit-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUIResetPasswordSuccessHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIResetPasswordSuccessHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="simple-form vertical-form form-fields-container">

<div class="title primary-txt">{{ localize "reset-password-success-page-title" }}</div>

{{ template "ERROR" . }}

<!-- FIXME: x_login_id -->

<div class="description primary-txt">{{ localize "reset-password-success-description" .x_login_id }}</div>

</div>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUISignupHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISignupHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
	<div class="content">
		{{ template "auth_ui_header.html" . }}
		<div class="authorize-form">
			<div class="authorize-idp-section">
				{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				<form class="authorize-idp-form" method="post" novalidate>
				{{ $.csrfField }}
				<button class="btn sso-btn {{ .provider_type }}" type="submit" name="x_idp_id" value="{{ .provider_alias }}" data-form-xhr="false">
					{{- if eq .provider_type "apple" -}}
					{{ localize "sign-up-apple" }}
					{{- end -}}
					{{- if eq .provider_type "google" -}}
					{{ localize "sign-up-google" }}
					{{- end -}}
					{{- if eq .provider_type "facebook" -}}
					{{ localize "sign-up-facebook" }}
					{{- end -}}
					{{- if eq .provider_type "linkedin" -}}
					{{ localize "sign-up-linkedin" }}
					{{- end -}}
					{{- if eq .provider_type "azureadv2" -}}
					{{ localize "sign-up-azureadv2" }}
					{{- end -}}
				</button>
				</form>
				{{ end }}
				{{ end }}
			</div>

			{{ $has_oauth := false }}
			{{ $has_login_id := false }}
			{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				{{ $has_oauth = true }}
				{{ end }}
				{{ if eq .type "login_id" }}
				{{ $has_login_id = true }}
				{{ end }}
			{{ end }}
			{{ if $has_oauth }}{{ if $has_login_id }}
			<div class="primary-txt sso-loginid-separator">{{ localize "sso-login-id-separator" }}</div>
			{{ end }}{{ end }}

			{{ template "ERROR" . }}

			<form class="authorize-loginid-form" method="post" novalidate>
				{{ $.csrfField }}
				<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">

				{{ range .x_identity_candidates }}
				{{ if eq .type "login_id" }}{{ if eq .login_id_key $.x_login_id_key }}
				{{ if eq .login_id_type "phone" }}
					<div class="phone-input">
						<select class="input select primary-txt" name="x_calling_code">
							{{ range $.x_calling_codes }}
							<option
								value="{{ . }}"
								{{ if $.x_calling_code }}{{ if eq $.x_calling_code . }}
								selected
								{{ end }}{{ end }}
								>
								+{{ . }}
							</option>
							{{ end }}
						</select>
						<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}">
					</div>
				{{ else }}
					<input class="input text-input primary-txt" type="{{ $.x_login_id_input_type }}" name="x_login_id" placeholder="{{ .login_id_type }}">
				{{ end }}
				{{ end }}{{ end }}
				{{ end }}

				{{ range .x_identity_candidates }}
				{{ if eq .type "login_id" }}{{ if not (eq .login_id_key $.x_login_id_key) }}
					<a class="link align-self-flex-start"
						href="{{ call $.MakeURLWithQuery "x_login_id_key" .login_id_key "x_login_id_input_type" .login_id_input_type}}">
						{{ localize "use-login-id-key" .login_id_key }}
					</a>
				{{ end }}{{ end }}
				{{ end }}

				<div class="link align-self-flex-start">
					<span class="primary-text">{{ localize "login-button-hint" }}</span>
					<a href="{{ call .MakeURLWithPathWithoutX "/login" }}">{{ localize "login-button-label" }}</a>
				</div>

				{{ if .x_password_authenticator_enabled }}
				<a class="link align-self-flex-start" href="{{ call .MakeURLWithPathWithoutX "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>
				{{ end }}

				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">
					{{ localize "next-button-label" }}
				</button>
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

var TemplateAuthUIPromoteHTML = template.Spec{
	Type:        TemplateItemTypeAuthUIPromoteHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
	<div class="content">
		{{ template "auth_ui_header.html" . }}
		<div class="authorize-form">
			<div class="authorize-idp-section">
				{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				<form class="authorize-idp-form" method="post" novalidate>
				{{ $.csrfField }}
				<button class="btn sso-btn {{ .provider_type }}" type="submit" name="x_idp_id" value="{{ .provider_alias }}" data-form-xhr="false">
					{{- if eq .provider_type "apple" -}}
					{{ localize "sign-up-apple" }}
					{{- end -}}
					{{- if eq .provider_type "google" -}}
					{{ localize "sign-up-google" }}
					{{- end -}}
					{{- if eq .provider_type "facebook" -}}
					{{ localize "sign-up-facebook" }}
					{{- end -}}
					{{- if eq .provider_type "linkedin" -}}
					{{ localize "sign-up-linkedin" }}
					{{- end -}}
					{{- if eq .provider_type "azureadv2" -}}
					{{ localize "sign-up-azureadv2" }}
					{{- end -}}
				</button>
				</form>
				{{ end }}
				{{ end }}
			</div>

			{{ $has_oauth := false }}
			{{ $has_login_id := false }}
			{{ range .x_identity_candidates }}
				{{ if eq .type "oauth" }}
				{{ $has_oauth = true }}
				{{ end }}
				{{ if eq .type "login_id" }}
				{{ $has_login_id = true }}
				{{ end }}
			{{ end }}
			{{ if $has_oauth }}{{ if $has_login_id }}
			<div class="primary-txt sso-loginid-separator">{{ localize "sso-login-id-separator" }}</div>
			{{ end }}{{ end }}

			{{ template "ERROR" . }}

			<form class="authorize-loginid-form" method="post" novalidate>
				{{ $.csrfField }}
				<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">

				{{ range .x_identity_candidates }}
				{{ if eq .type "login_id" }}{{ if eq .login_id_key $.x_login_id_key }}
				{{ if eq .login_id_type "phone" }}
					<div class="phone-input">
						<select class="input select primary-txt" name="x_calling_code">
							{{ range $.x_calling_codes }}
							<option
								value="{{ . }}"
								{{ if $.x_calling_code }}{{ if eq $.x_calling_code . }}
								selected
								{{ end }}{{ end }}
								>
								+{{ . }}
							</option>
							{{ end }}
						</select>
						<input class="input text-input primary-txt" type="text" inputmode="numeric" pattern="[0-9]*" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}">
					</div>
				{{ else }}
					<input class="input text-input primary-txt" type="{{ $.x_login_id_input_type }}" name="x_login_id" placeholder="{{ .login_id_type }}">
				{{ end }}
				{{ end }}{{ end }}
				{{ end }}

				{{ range .x_identity_candidates }}
				{{ if eq .type "login_id" }}{{ if not (eq .login_id_key $.x_login_id_key) }}
					<a class="link align-self-flex-start"
						href="{{ call $.MakeURLWithQuery "x_login_id_key" .login_id_key "x_login_id_input_type" .login_id_input_type}}">
						{{ localize "use-login-id-key" .login_id_key }}
					</a>
				{{ end }}{{ end }}
				{{ end }}

				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">
					{{ localize "next-button-label" }}
				</button>
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

var TemplateAuthUICreatePasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUICreatePasswordHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<form class="simple-form vertical-form form-fields-container" method="post" novalidate>
{{ $.csrfField }}

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ "back-button-title" }}"></button>
	<div class="login-id primary-txt">
	{{ if .x_national_number }}
		+{{ .x_calling_code}} {{ .x_national_number }}
	{{ else }}
		{{ .x_login_id }}
	{{ end }}
	</div>
</div>

<div class="title primary-txt">{{ localize "create-password-page-title" }}</div>

{{ template "ERROR" . }}

<input id="password" data-password-policy-password="" class="input text-input primary-txt" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

{{ template "PASSWORD_POLICY" . }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "next-button-label" }}</button>

</form>
{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUISettingsHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISettingsHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="settings-form primary-txt">
  You are authenticated. To logout, please visit <a href="/logout">here</a>.
</div>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUISettingsIdentityHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISettingsIdentityHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<div class="settings-identity">
  <h1 class="title primary-txt">{{ localize "settings-identity-title" }}</h1>

  {{ template "ERROR" . }}

  {{ range .x_identity_candidates }}
  <div class="identity">
    <div class="icon {{ .type }} {{ .provider_type }} {{ .login_id_type }}"></div>
    <div class="identity-info flex-child-no-overflow">
      <h2 class="identity-name primary-txt">
         {{ if eq .type "oauth" }}
           {{ if eq .provider_type "google" }}
           {{ localize "settings-identity-oauth-google" }}
           {{ end }}
           {{ if eq .provider_type "apple" }}
           {{ localize "settings-identity-oauth-apple" }}
           {{ end }}
           {{ if eq .provider_type "facebook" }}
           {{ localize "settings-identity-oauth-facebook" }}
           {{ end }}
           {{ if eq .provider_type "linkedin" }}
           {{ localize "settings-identity-oauth-linkedin" }}
           {{ end }}
           {{ if eq .provider_type "azureadv2" }}
           {{ localize "settings-identity-oauth-azureadv2" }}
           {{ end }}
         {{ end }}
         {{ if eq .type "login_id" }}
           {{ if eq .login_id_type "email" }}
           {{ localize "settings-identity-login-id-email" }}
           {{ end }}
           {{ if eq .login_id_type "phone" }}
           {{ localize "settings-identity-login-id-phone" }}
           {{ end }}
           {{ if eq .login_id_type "username" }}
           {{ localize "settings-identity-login-id-username" }}
           {{ end }}
           {{ if eq .login_id_type "raw" }}
           {{ localize "settings-identity-login-id-raw" }}
           {{ end }}
         {{ end }}
      </h2>

      {{ if eq .type "oauth" }}{{ if .email }}
      <h3 class="identity-claim secondary-txt text-ellipsis">
        {{ .email }}
      </h3>
      {{ end }}{{ end }}

      {{ if eq .type "login_id" }}{{ if .login_id_value }}
      <h3 class="identity-claim secondary-txt text-ellipsis">
        {{ .login_id_value }}
      </h3>
      {{ end }}{{ end }}
    </div>

    {{ if eq .type "oauth" }}
      <form method="post" novalidate>
      {{ $.csrfField }}
      <input type="hidden" name="x_idp_id" value="{{ .provider_alias }}">
      {{ if .provider_subject_id }}
      <button class="btn destructive-btn" type="submit" name="x_action" value="unlink">{{ localize "disconnect-button-label" }}</button>
      {{ else }}
      <button class="btn primary-btn" type="submit" name="x_action" value="link" data-form-xhr="false">{{ localize "connect-button-label" }}</button>
      {{ end }}
      </form>
    {{ end }}

    {{ if eq .type "login_id" }}
      <form method="post" novalidate>
      {{ $.csrfField }}
      <input type="hidden" name="x_login_id_key" value="{{ .login_id_key }}">
      <input type="hidden" name="x_login_id_type" value="{{ .login_id_type }}">
      {{ if eq .login_id_type "phone" }}
      <input type="hidden" name="x_login_id_input_type" value="phone">
      {{ else if eq .login_id_type "email" }}
      <input type="hidden" name="x_login_id_input_type" value="email">
      {{ else }}
      <input type="hidden" name="x_login_id_input_type" value="text">
      {{ end }}
      {{ if .login_id_value }}
      <input type="hidden" name="x_old_login_id_value" value="{{ .login_id_value }}">
      <button class="btn secondary-btn" type="submit" name="x_action" value="login_id">{{ localize "change-button-label" }}</a>
      {{ else }}
      <button class="btn primary-btn" type="submit" name="x_action" value="login_id">{{ localize "connect-button-label" }}</a>
      {{ end }}
      </form>
    {{ end }}
  </div>
  {{ end }}
</div>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUILogoutHTML = template.Spec{
	Type:        TemplateItemTypeAuthUILogoutHTML,
	IsHTML:      true,
	Translation: TemplateItemTypeAuthUITranslationJSON,
	Defines:     defines,
	Components:  components,
	Default: `<!DOCTYPE html>
<html>
{{ template "auth_ui_html_head.html" . }}
<body class="page">
<div class="content">

{{ template "auth_ui_header.html" . }}

<form class="logout-form" method="post" novalidate>
  {{ $.csrfField }}
  <p class="primary-txt">{{ localize "logout-button-hint" }}</p>
  <button class="btn primary-btn align-self-center" type="submit" name="x_action" value="logout">{{ localize "logout-button-label" }}</button>
</form>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}
