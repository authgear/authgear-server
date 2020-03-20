package webapp

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeAuthUISignInHTML config.TemplateItemType = "auth_ui_sign_in.html"
	// nolint: gosec
	TemplateItemTypeAuthUISignInPasswordHTML config.TemplateItemType = "auth_ui_sign_in_password.html"
	TemplateItemTypeAuthUISignUpHTML         config.TemplateItemType = "auth_ui_sign_up.html"
	// nolint: gosec
	TemplateItemTypeAuthUISignUpPasswordHTML config.TemplateItemType = "auth_ui_sign_up_password.html"
	TemplateItemTypeAuthUISettingsHTML       config.TemplateItemType = "auth_ui_settings.html"
)

const defineHead = `
{{ define "HEAD" }}
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
`

const defineLogo = `
{{ define "LOGO" }}
{{ if .logo_uri }}
<div class="logo" style="background-image: url('{{ .logo_uri }}'); background-position: center; background-size: contain; background-repeat: no-repeat"></div>
{{ else }}
<div class="logo"></div>
{{ end }}
{{ end }}
`

const defineError = `
{{ define "ERROR" }}
{{ if .x_error }}
<ul class="errors">
	{{ if eq .x_error.reason "ValidationFailed" }}
		{{ range .x_error.info.causes }}
		{{ if and (eq .kind "Required") (eq .pointer "/x_login_id" ) }}
		<li class="error-txt">Email or Username is required</li>
		{{ else if and (eq .kind "Required") (eq .pointer "/x_calling_code" ) }}
		<li class="error-txt">Calling code is required</li>
		{{ else if and (eq .kind "Required") (eq .pointer "/x_national_number" ) }}
		<li class="error-txt">Phone number is required</li>
		{{ else if and (eq .kind "StringFormat") (eq .pointer "/x_national_number" ) }}
		<li class="error-txt">Phone number must contain digits only</li>
		{{ else if and (eq .kind "StringFormat") (eq .pointer "/login_ids/0/value") }}
		<li class="error-txt">Invalid email address</li>
		{{ else }}
		<li class="error-txt">{{ .message }}</li>
		{{ end }}
		{{ end }}
	{{ else if eq .x_error.reason "InvalidCredentials" }}
		<li class="error-txt">Incorrect email, phone number, username, or password</li>
	{{ else }}
		<li class="error-txt">{{ .x_error.message }}</li>
	{{ end }}
</ul>
{{ end }}
{{ end }}
`

const defineSkygearLogo = `
{{ define "SKYGEAR_LOGO" }}
<div class="skygear-logo"></div>
{{ end }}
`

var defines = []string{
	defineHead,
	defineLogo,
	defineError,
	defineSkygearLogo,
}

var TemplateAuthUISignInHTML = template.Spec{
	Type:    TemplateItemTypeAuthUISignInHTML,
	IsHTML:  true,
	Defines: defines,
	Default: `<!DOCTYPE html>
<html>
{{ template "HEAD" . }}
<body class="page">
	<div class="content">
		{{ template "LOGO" . }}
		<div class="authorize-form">
			<form class="authorize-idp-form" method="post">
				<input type="hidden" name="x_step" value="choose_idp">
				{{ range .x_idp_providers }}
				<button class="btn sso-btn {{ .type }}" type="submit" name="x_idp_id" value="{{ .id }}">
					{{- if eq .type "apple" -}}
					Sign in with Apple
					{{- end -}}
					{{- if eq .type "google" -}}
					Sign in with Google
					{{- end -}}
					{{- if eq .type "facebook" -}}
					Sign in with Facebook
					{{- end -}}
					{{- if eq .type "instagram" -}}
					Sign in with Instagram
					{{- end -}}
					{{- if eq .type "linkedin" -}}
					Sign in with LinkedIn
					{{- end -}}
					{{- if eq .type "azureadv2" -}}
					Sign in with Azure AD
					{{- end -}}
				</button>
				{{ end }}
			</form>

			<div class="primary-txt sso-loginid-separator">or</div>

			{{ template "ERROR" . }}

			<form id="empty-form" method="post"></form>

			<form class="authorize-loginid-form" method="post">
				<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">

				{{ if .x_login_id_input_type }}{{ if and (eq .x_login_id_input_type "phone") .x_login_id_input_type_has_phone }}
				<div class="phone-input">
					<select class="input select" name="x_calling_code">
						<option value="">Code</option>
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
					<input class="input text-input" type="tel" name="x_national_number" placeholder="Phone number" value="{{ .x_national_number }}">
				</div>
				{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if and (not (eq .x_login_id_input_type "phone")) .x_login_id_input_type_has_text }}
				<input class="input text-input" type="text" name="x_login_id" placeholder="Email or Username" value="{{ .x_login_id }}">
				{{ end }}{{ end }}

				{{ if .x_login_id_input_type }}{{ if and (eq .x_login_id_input_type "phone") .x_login_id_input_type_has_text }}
				<button class="link anchor" type="submit" name="x_login_id_input_type" value="text" form="empty-form">Use an email or username instead</button>
				{{ end }}{{ end }}
				{{ if .x_login_id_input_type }}{{ if and (not (eq .x_login_id_input_type "phone")) .x_login_id_input_type_has_phone }}
				<button class="link anchor" type="submit" name="x_login_id_input_type" value="phone" form="empty-form">Use a phone number instead</button>
				{{ end }}{{ end }}

				<div class="link">
					<span class="primary-text">Don't have an account yet? </span>
					<button type="submit" class="anchor" name="x_step" value="sign_up" form="empty-form">Create one!</button>
				</div>
				<a class="link anchor" href="#">Can't access your account?</a>

				{{ if or .x_login_id_input_type_has_phone .x_login_id_input_type_has_text }}
				<button class="btn primary-btn" type="submit" name="x_step" value="submit_login_id">Next</button>
				{{ end }}
			</form>
		</div>
		{{ template "SKYGEAR_LOGO" . }}
	</div>
</body>
</html>
`,
}

var TemplateAuthUISignInPasswordHTML = template.Spec{
	Type:    TemplateItemTypeAuthUISignInPasswordHTML,
	IsHTML:  true,
	Defines: defines,
	Default: `<!DOCTYPE html>
<html>
{{ template "HEAD" . }}
<body class="page">
<div class="content">

{{ template "LOGO" . }}

<form class="enter-password-form" method="post">

<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">

<div class="nav-bar">
	<button class="btn back-btn" onclick="window.history.back()" title="Back"></button>
	<div class="login-id primary-txt">
	{{ if .x_calling_code }}
		+{{ .x_calling_code}} {{ .x_national_number }}
	{{ else }}
		{{ .x_login_id }}
	{{ end }}
	</div>
</div>

<div class="title primary-txt">Enter password</div>

{{ template "ERROR" . }}

<input type="hidden" name="x_calling_code" value="{{ .x_calling_code }}">
<input type="hidden" name="x_national_number" value="{{ .x_national_number }}">
<input type="hidden" name="x_login_id" value="{{ .x_login_id }}">

<input id="password" class="input text-input" type="password" name="x_password" placeholder="Password" value="{{ .x_password }}">

<button class="btn secondary-btn toggle-password-visibility"></button>

<a class="anchor" href="">Forgot Password?</a>

<button class="btn primary-btn" type="submit" name="x_step" value="submit_password">Next</button>

</form>
{{ template "SKYGEAR_LOGO" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUISignUpHTML = template.Spec{
	Type:    TemplateItemTypeAuthUISignUpHTML,
	IsHTML:  true,
	Defines: defines,
	Default: `<!DOCTYPE html>
<html>
{{ template "HEAD" . }}
<body class="page">
	<div class="content">
		{{ template "LOGO" . }}
		<div class="authorize-form">
			{{ template "ERROR" . }}

			<form id="empty-form" method="post"></form>

			{{ range .x_login_id_keys }}
			<form id="sign_up-{{ .key }}" method="post">
				<input type="hidden" name="x_step" value="sign_up">
				{{ if eq .type "phone" }}
					<input type="hidden" name="x_login_id_input_type" value="phone">
				{{ else }}
					<input type="hidden" name="x_login_id_input_type" value="text">
				{{ end }}
			</form>
			{{ end }}

			<form class="authorize-loginid-form" method="post">
				<input type="hidden" name="x_login_id_key" value="{{ .x_login_id_key }}">
				<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">

				{{ range .x_login_id_keys }}
					{{ if eq .key $.x_login_id_key }}
					{{ if eq .type "phone" }}
					<div class="phone-input">
						<select class="input select" name="x_calling_code">
							<option value="">Code</option>
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
						<input class="input text-input" type="tel" name="x_national_number" placeholder="Phone number" value="{{ $.x_national_number }}">
					</div>
					{{ else }}
					<input class="input text-input" type="text" name="x_login_id" placeholder="{{ .type }}" value="{{ $.x_login_id }}">
					{{ end }}
					{{ end }}
				{{ end }}

				{{ range .x_login_id_keys }}
					{{ if not (eq .key $.x_login_id_key) }}
					<button class="link anchor" type="submit" name="x_login_id_key" value="{{ .key }}" form="sign_up-{{ .key }}">Use {{ .key }} instead</button>
					{{ end }}
				{{ end }}

				<div class="link">
					<span class="primary-text">Have an account already? </span>
					<button type="submit" class="anchor" name="x_step" value="" form="empty-form">Sign in!</button>
				</div>
				<a class="link anchor" href="#">Can't access your account?</a>

				<button class="btn primary-btn" type="submit" name="x_step" value="sign_up_submit_login_id">Next</button>
			</form>
		</div>
		{{ template "SKYGEAR_LOGO" . }}
	</div>
</body>
</html>
`,
}

var TemplateAuthUISignUpPasswordHTML = template.Spec{
	Type:    TemplateItemTypeAuthUISignUpPasswordHTML,
	IsHTML:  true,
	Defines: defines,
	Default: `<!DOCTYPE html>
<html>
{{ template "HEAD" . }}
<body class="page">
<div class="content">

{{ template "LOGO" . }}

TODO(webapp): sign in password page

{{ template "SKYGEAR_LOGO" . }}

</div>
</body>
</html>
`,
}

var TemplateAuthUISettingsHTML = template.Spec{
	Type:    TemplateItemTypeAuthUISettingsHTML,
	IsHTML:  true,
	Defines: defines,
	Default: `<!DOCTYPE html>
<html>
{{ template "HEAD" . }}
<body class="page">
<div class="content">

{{ template "LOGO" . }}

You are authenticated. To logout, please clear the cookie AND revisit this page. Refreshing causes the form to be submitted again so you will become authenticated again.

{{ template "SKYGEAR_LOGO" . }}

</div>
</body>
</html>
`,
}
