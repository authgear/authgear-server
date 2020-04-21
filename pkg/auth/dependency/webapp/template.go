package webapp

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeAuthUIHTMLHeadHTML config.TemplateItemType = "auth_ui_html_head.html"
	TemplateItemTypeAuthUIHeaderHTML   config.TemplateItemType = "auth_ui_header.html"
	TemplateItemTypeAuthUIFooterHTML   config.TemplateItemType = "auth_ui_footer.html"

	TemplateItemTypeAuthUILoginHTML config.TemplateItemType = "auth_ui_login.html"
	// nolint: gosec
	TemplateItemTypeAuthUILoginPasswordHTML config.TemplateItemType = "auth_ui_login_password.html"
	TemplateItemTypeAuthUISignupHTML        config.TemplateItemType = "auth_ui_signup.html"
	// nolint: gosec
	TemplateItemTypeAuthUISignupPasswordHTML config.TemplateItemType = "auth_ui_signup_password.html"
	// nolint: gosec
	TemplateItemTypeAuthUIForgotPasswordHTML config.TemplateItemType = "auth_ui_forgot_password.html"
	// nolint: gosec
	TemplateItemTypeAuthUIForgotPasswordSuccessHTML config.TemplateItemType = "auth_ui_forgot_password_success.html"
	// nolint: gosec
	TemplateItemTypeAuthUIResetPasswordHTML config.TemplateItemType = "auth_ui_reset_password.html"
	TemplateItemTypeAuthUISettingsHTML      config.TemplateItemType = "auth_ui_settings.html"
	TemplateItemTypeAuthUILogoutHTML        config.TemplateItemType = "auth_ui_logout.html"
)

var TemplateAuthUIHTMLHeadHTML = template.Spec{
	Type:   TemplateItemTypeAuthUIHTMLHeadHTML,
	IsHTML: true,
	Default: `
{{ define "auth_ui_html_head.html" }}
<head>
<title>{{ .client_name }}</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="{{ .x_static_asset_url_prefix }}/css/main.css">
<script src="{{ .x_static_asset_url_prefix}}/js/main.js"></script>
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
<div class="skygear-logo"></div>
{{ end }}
`,
}

const defineError = `
{{ define "ERROR" }}
{{ if .x_error }}
<ul class="errors">
	{{ if eq .x_error.reason "ValidationFailed" }}
		{{ range .x_error.info.causes }}
		{{ if and (eq .kind "Required") (eq .pointer "/x_login_id" ) }}
		<li class="error-txt">{{ localize "error-email-or-username-required" }}</li>
		{{ else if and (eq .kind "Required") (eq .pointer "/x_password" ) }}
		<li class="error-txt">{{ localize "error-password-required" }}</li>
		{{ else if and (eq .kind "Required") (eq .pointer "/x_calling_code" ) }}
		<li class="error-txt">{{ localize "error-calling-code-required" }}</li>
		{{ else if and (eq .kind "Required") (eq .pointer "/x_national_number" ) }}
		<li class="error-txt">{{ localize "error-phone-number-required" }}</li>
		{{ else if and (eq .kind "StringFormat") (eq .pointer "/x_national_number" ) }}
		<li class="error-txt">{{ localize "error-phone-number-format" }}</li>
		{{ else if and (eq .kind "StringFormat") (eq .pointer "/login_ids/0/value") }}
			{{ range $.x_login_id_keys }}
				{{ if eq .key $.x_login_id_key }}
					{{ if eq .type "email" }}
					<li class="error-txt">{{ localize "error-invalid-email" }}</li>
					{{ else }}
					<li class="error-txt">{{ localize "error-invalid-username" }}</li>
					{{ end }}
				{{ end }}
			{{ end }}
		{{ else }}
		<li class="error-txt">{{ .message }}</li>
		{{ end }}
		{{ end }}
	{{ else if eq .x_error.reason "InvalidCredentials" }}
		<li class="error-txt">{{ localize "error-invalid-credentials" }}</li>
	{{ else if eq .x_error.reason "PasswordPolicyViolated" }}
		<!-- This error is handled differently -->
	{{ else }}
		<li class="error-txt">{{ .x_error.message }}</li>
	{{ end }}
</ul>
{{ end }}
{{ end }}
`

const definePasswordPolicy = `
{{ define "PASSWORD_POLICY" }}
{{ if .x_password_policies }}
<ul>
{{ range .x_password_policies }}
  {{ if eq .kind "PasswordTooShort" }}
  <li class="password-policy length {{ template "PASSWORD_POLICY_CLASS" . }}" data-min-length="{{ .min_length}}">
    {{ localize "password-policy-minimum-length" .min_length }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordUppercaseRequired" }}
  <li class="password-policy uppercase {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-uppercase" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordLowercaseRequired" }}
  <li class="password-policy lowercase {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-lowercase" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordDigitRequired" }}
  <li class="password-policy digit {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-digit" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordSymbolRequired" }}
  <li class="password-policy symbol {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-symbol" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordContainingExcludedKeywords" }}
  <li class="password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
    {{ localize "password-policy-banned-words" }}
  </li>
  {{ end }}
  {{ if eq .kind "PasswordBelowGuessableLevel" }}
    {{ if eq .min_level 1.0 }}
    <li class="password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-1" }}
    </li>
    {{ end }}
    {{ if eq .min_level 2.0 }}
    <li class="password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-2" }}
    </li>
    {{ end }}
    {{ if eq .min_level 3.0 }}
    <li class="password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-3" }}
    </li>
    {{ end }}
    {{ if eq .min_level 4.0 }}
    <li class="password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
      {{ localize "password-policy-guessable-level-4" }}
    </li>
    {{ end }}
    {{ if eq .min_level 5.0 }}
    <li class="password-policy {{ template "PASSWORD_POLICY_CLASS" . }}">
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
			<form class="authorize-idp-form" method="post">
				{{ $.csrfField }}
				{{ range .x_idp_providers }}
				<button class="btn sso-btn {{ .type }}" type="submit" name="x_idp_id" value="{{ .id }}">
					{{- if eq .type "apple" -}}
					{{ localize "sign-in-apple" }}
					{{- end -}}
					{{- if eq .type "google" -}}
					{{ localize "sign-in-google" }}
					{{- end -}}
					{{- if eq .type "facebook" -}}
					{{ localize "sign-in-facebook" }}
					{{- end -}}
					{{- if eq .type "instagram" -}}
					{{ localize "sign-in-instagram" }}
					{{- end -}}
					{{- if eq .type "linkedin" -}}
					{{ localize "sign-in-linkedin" }}
					{{- end -}}
					{{- if eq .type "azureadv2" -}}
					{{ localize "sign-in-azureadv2" }}
					{{- end -}}
				</button>
				{{ end }}
			</form>

			{{ if .x_idp_providers }}{{ if or .x_login_id_input_type_has_phone .x_login_id_input_type_has_text }}
			<div class="primary-txt sso-loginid-separator">{{ localize "sso-login-id-separator" }}</div>
			{{ end }}{{ end }}

			{{ template "ERROR" . }}

			<form class="authorize-loginid-form" method="post">
				{{ $.csrfField }}

				{{ if .x_login_id_input_type }}{{ if and (eq .x_login_id_input_type "phone") .x_login_id_input_type_has_phone }}
				<div class="phone-input">
					<select class="input select" name="x_calling_code">
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
					<input class="input text-input" type="tel" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}" value="{{ .x_national_number }}">
				</div>
				{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if and (not (eq .x_login_id_input_type "phone")) .x_login_id_input_type_has_text }}
				<input class="input text-input" type="text" name="x_login_id" placeholder="{{ localize "login-id-placeholder" }}" value="{{ .x_login_id }}">
				{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if and (eq .x_login_id_input_type "phone") .x_login_id_input_type_has_text }}
				<a class="link anchor align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "text" }}">{{ localize "use-text-login-id-description" }}</a>
				{{ end }}{{ end }}
				{{ if .x_login_id_input_type }}{{ if and (not (eq .x_login_id_input_type "phone")) .x_login_id_input_type_has_phone }}
				<a class="link anchor align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
				{{ end }}{{ end }}

				<div class="link">
					<span class="primary-text">{{ localize "signup-button-hint" }}</span>
					<a class="anchor" href="{{ call .MakeURLWithPath "/signup" }}">{{ localize "signup-button-label" }}</a>
				</div>
				<a class="link anchor align-self-flex-start" href="{{ call .MakeURLWithPath "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>

				{{ if or .x_login_id_input_type_has_phone .x_login_id_input_type_has_text }}
				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "confirm-login-id-button-label" }}</button>
				{{ end }}
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

var TemplateAuthUILoginPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUILoginPasswordHTML,
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

<form class="enter-password-form" method="post">
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

<input type="hidden" name="x_calling_code" value="{{ .x_calling_code }}">
<input type="hidden" name="x_national_number" value="{{ .x_national_number }}">
<input type="hidden" name="x_login_id" value="{{ .x_login_id }}">

<input id="password" class="input text-input" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}" value="{{ .x_password }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

<a class="anchor link align-self-flex-start" href="{{ call .MakeURLWithPath "/forgot_password" }}">{{ localize "forgot-password-button-label--enter-password-page" }}</a>

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "confirm-password-button-label" }}</button>

</form>
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

<form class="forgot-password-form" method="post">
{{ $.csrfField }}

<div class="nav-bar">
	<button class="btn back-btn" type="button" title="{{ localize "back-button-title" }}"></button>
</div>

<div class="title primary-txt">{{ localize "forgot-password-page-title" }}</div>

{{ template "ERROR" . }}

{{ if .x_login_id_input_type }}{{ if and (eq .x_login_id_input_type "phone") .x_login_id_input_type_has_phone }}
<div class="description primary-txt">{{ localize "forgot-password-phone-description" }}</div>
<div class="phone-input">
	<select class="input select" name="x_calling_code">
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
	<input class="input text-input" type="tel" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}" value="{{ .x_national_number }}">
</div>
{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if and (not (eq .x_login_id_input_type "phone")) .x_login_id_input_type_has_text }}
<div class="description primary-txt">{{ localize "forgot-password-email-description" }}</div>
<input class="input text-input" type="text" name="x_login_id" placeholder="{{ localize "email-placeholder" }}" value="{{ .x_login_id }}">
{{ end }}{{ end }}

{{ if .x_login_id_input_type }}{{ if and (eq .x_login_id_input_type "phone") .x_login_id_input_type_has_text }}
<a class="link anchor align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "text" }}">{{ localize "use-email-login-id-description" }}</a>
{{ end }}{{ end }}
{{ if .x_login_id_input_type }}{{ if and (not (eq .x_login_id_input_type "phone")) .x_login_id_input_type_has_phone }}
<a class="link anchor align-self-flex-start" href="{{ call .MakeURLWithQuery "x_login_id_input_type" "phone" }}">{{ localize "use-phone-login-id-description" }}</a>
{{ end }}{{ end }}

{{ if or .x_login_id_input_type_has_phone .x_login_id_input_type_has_text }}
<button class="btn primary-btn submit-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "confirm-login-id-button-label" }}</button>
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

<div class="forgot-password-success">

<div class="title primary-txt">{{ localize "forgot-password-success-page-title" }}</div>

{{ template "ERROR" . }}

<div class="description primary-txt">{{ localize "forgot-password-success-description" .x_login_id }}</div>

<a class="anchor btn primary-btn align-self-flex-end" href="{{ call .MakeURLWithPath "/login" }}">{{ localize "login-button-label--forgot-password-success-page" }}</a>

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

<form class="reset-password-form" method="post">
{{ $.csrfField }}

<div class="title primary-txt">{{ localize "reset-password-page-title" }}</div>

{{ template "ERROR" . }}

<div class="description primary-txt">{{ localize "reset-password-description" }}</div>

<input id="password" data-password-policy-password="" class="input text-input" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}" value="{{ .x_password }}">

<button class="btn secondary-btn password-visibility-btn show-password">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password">{{ localize "hide-password" }}</button>

{{ template "PASSWORD_POLICY" . }}

<button class="btn primary-btn submit-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "confirm-password-button-label" }}</button>

</form>

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
			{{ template "ERROR" . }}

			<form class="authorize-loginid-form" method="post">
				{{ $.csrfField }}
				<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">

				{{ range .x_login_id_keys }}
					{{ if eq .key $.x_login_id_key }}
					{{ if eq .type "phone" }}
					<div class="phone-input">
						<select class="input select" name="x_calling_code">
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
						<input class="input text-input" type="tel" name="x_national_number" placeholder="{{ localize "phone-number-placeholder" }}" value="{{ $.x_national_number }}">
					</div>
					{{ else }}
					<input class="input text-input" type="text" name="x_login_id" placeholder="{{ .type }}" value="{{ $.x_login_id }}">
					{{ end }}
					{{ end }}
				{{ end }}

				{{ range .x_login_id_keys }}
					{{ if not (eq .key $.x_login_id_key) }}
					<a class="link anchor align-self-flex-start"
						href="{{ call $.MakeURLWithQuery "x_login_id_key" .key "x_login_id_input_type" .input_type}}">
						{{ localize "use-login-id-key" .key }}
					</a>
					{{ end }}
				{{ end }}

				<div class="link align-self-flex-start">
					<span class="primary-text">{{ localize "login-button-hint" }}</span>
					<a class="anchor" href="{{ call .MakeURLWithPath "/login" }}">{{ localize "login-button-label" }}<a>
				</div>
				<a class="link anchor align-self-flex-start" href="{{ call .MakeURLWithPath "/forgot_password" }}">{{ localize "forgot-password-button-label" }}</a>

				<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">
					{{ localize "confirm-login-id-button-label" }}
				</button>
			</form>
		</div>
		{{ template "auth_ui_footer.html" . }}
	</div>
</body>
</html>
`,
}

var TemplateAuthUISignupPasswordHTML = template.Spec{
	Type:        TemplateItemTypeAuthUISignupPasswordHTML,
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

<form class="enter-password-form" method="post">
{{ $.csrfField }}
<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">
<input type="hidden" name="x_calling_code" value="{{ .x_calling_code }}">
<input type="hidden" name="x_national_number" value="{{ .x_national_number }}">
<input type="hidden" name="x_login_id" value="{{ .x_login_id }}">

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

<input id="password" data-password-policy-password="" class="input text-input" type="password" name="x_password" placeholder="{{ localize "password-placeholder" }}" value="{{ .x_password }}">

<button class="btn secondary-btn password-visibility-btn show-password" type="button">{{ localize "show-password" }}</button>
<button class="btn secondary-btn password-visibility-btn hide-password" type="button">{{ localize "hide-password" }}</button>

{{ template "PASSWORD_POLICY" . }}

<button class="btn primary-btn align-self-flex-end" type="submit" name="submit" value="">{{ localize "confirm-password-button-label" }}</button>

{{ if eq .x_login_id_input_type "phone" }}
<p class="description">
{{ localize "sms-charge-warning" }}
</p>
{{ end }}

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

<div class="settings-form">
  You are authenticated. To logout, please visit <a href="/logout">here</a>.
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

<form class="logout-form" method="post">
  {{ $.csrfField }}
  <p>{{ localize "logout-button-hint" }}</p>
  <button class="btn primary-btn align-self-center" type="submit" name="x_action" value="logout">{{ localize "logout-button-label" }}</button>
</form>

{{ template "auth_ui_footer.html" . }}

</div>
</body>
</html>
`,
}
