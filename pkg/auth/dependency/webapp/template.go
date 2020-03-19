package webapp

import (
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/template"
)

const (
	TemplateItemTypeAuthUISignInHTML config.TemplateItemType = "auth_ui_sign_in.html"
	// nolint
	TemplateItemTypeAuthUISignInPasswordHTML config.TemplateItemType = "auth_ui_sign_in_password.html"
)

const defineHead = `
{{ define "HEAD" }}
<head>
<title>{{ .client_name }}</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<link rel="stylesheet" href="{{ .x_static_asset_url_prefix }}/css/main.css">
{{ if .x_css }}
<style>
{{ .x_css }}
</style>
{{ end }}
</head>
{{ end }}
`

const defineHidden = `
{{ define "HIDDEN" }}
<input type="hidden" name="x_login_id_input_type" value="{{ .x_login_id_input_type }}">
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
{{ if .x_error }}{{ if eq .x_error.reason "ValidationFailed" }}
<ul class="errors">
{{ range .x_error.info.causes }}
<li class="error-txt">{{ .message }}</li>
{{ end }}
</ul>
{{ else }}
<ul>
<li class="error-txt">{{ .x_error.message }}</li>
</ul>
{{ end }}{{ end }}
{{ end }}
`

const defineSkygearLogo = `
{{ define "SKYGEAR_LOGO" }}
<div class="skygear-logo"></div>
{{ end }}
`

var defines = []string{
	defineHead,
	defineHidden,
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

			<form class="authorize-loginid-form" method="post">
				{{ template "HIDDEN" . }}

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
				<a class="link anchor" href="{{ .x_use_text_url }}">Use an email or username instead</a>
				{{ end }}{{ end }}
				{{ if .x_login_id_input_type }}{{ if and (not (eq .x_login_id_input_type "phone")) .x_login_id_input_type_has_phone }}
				<a class="link anchor" href="{{ .x_use_phone_url }}">Use a phone number instead</a>
				{{ end }}{{ end }}

				<div class="link"><span class="primary-text">Don't have an account yet? </span><a class="anchor" href="#">Create one!</a></div>
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

{{ template "HIDDEN" . }}

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

<a class="anchor" href="">Forgot Password?</a>

<button class="btn primary-btn" type="submit" name="x_step" value="submit_password">Next</button>

</form>
{{ template "SKYGEAR_LOGO" . }}

</div>
</body>
</html>
`,
}
