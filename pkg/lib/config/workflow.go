package config

var _ = Schema.Add("WorkflowConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {}
}
`)

var _ = Schema.Add("WorkflowObjectID", `
{
	"type": "string",
	"pattern": "^[a-zA-Z_][a-zA-Z0-9_]*$"
}
`)

var _ = Schema.Add("WorkflowIdentificationMethod", `
{
	"type": "string",
	"enum": [
		"email",
		"phone",
		"username",
		"oauth",
		"passkey",
		"siwe"
	]
}
`)

var _ = Schema.Add("WorkflowAuthenticationMethod", `
{
	"type": "string",
	"enum": [
		"primary_password",
		"primary_passkey",
		"primary_oob_otp_email",
		"primary_oob_otp_sms",
		"secondary_password",
		"secondary_totp",
		"secondary_oob_otp_email",
		"secondary_oob_otp_sms"
	]
}
`)
