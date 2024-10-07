package testrunner

import (
	"net/http"
	"strings"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
)

type BeforeHook struct {
	Type       BeforeHookType      `json:"type"`
	UserImport string              `json:"user_import"`
	CustomSQL  BeforeHookCustomSQL `json:"custom_sql"`
}

var _ = TestCaseSchema.Add("AuthgearYAMLSource", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"extend": { "type": "string", "description": "Path to the base authgear.yaml" },
		"override": { "type": "string", "description": "Inline snippet to override the base authgear.yaml" }
	}
}
`)

type AuthgearYAMLSource struct {
	Extend   string `json:"extend"`
	Override string `json:"override"`
}

type BeforeHookType string

const (
	BeforeHookTypeUserImport BeforeHookType = "user_import"
	BeforeHookTypeCustomSQL  BeforeHookType = "custom_sql"
)

var _ = TestCaseSchema.Add("BeforeHookCustomSQL", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"path": { "type": "string", "description": "Path to the custom SQL script" }
	},
	"required": ["path"]
}
`)

type BeforeHookCustomSQL struct {
	Path string `json:"path"`
}

var _ = TestCaseSchema.Add("BeforeHook", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "type": "string", "enum": ["user_import", "custom_sql"] },
		"user_import": { "type": "string" },
		"custom_sql": { "$ref": "#/$defs/BeforeHookCustomSQL" }
	},
	"required": ["type"],
	"allOf": [
			{
				"if": { "properties": { "type": { "const": "user_import" } } },
				"then": {
					"required": ["user_import"]
				}
			},
			{
				"if": { "properties": { "type": { "const": "custom_sql" } } },
				"then": {
					"required": ["custom_sql"]
				}
			}
		]
}
`)

var _ = TestCaseSchema.Add("SAMLBinding", `
{
	"type": "string",
	"enum": [
		"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
		"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
	]
}
`)

var _ = TestCaseSchema.Add("Step", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string" },
		"action": { "type": "string", "enum": [
			"create",
			"input",
			"oauth_redirect",
			"generate_totp_code",
			"query",
			"saml_request",
			"http_request"
		]},
		"input": { "type": "string" },
		"to": { "type": "string" },
		"redirect_uri": { "type": "string" },
		"totp_secret": { "type": "string" },
		"output": { "$ref": "#/$defs/Output" },
		"query": { "type": "string" },
		"query_output": { "$ref": "#/$defs/QueryOutput" },
		"saml_output": { "$ref": "#/$defs/SAMLOutput" },
		"saml_request": { "type": "string" },
		"saml_request_destination": { "type": "string" },
		"saml_request_binding": { "$ref": "#/$defs/SAMLBinding" },
		"http_request_method": { "type": "string" },
		"http_request_url": { "type": "string" },
		"http_request_headers": { "type": "object" },
		"http_request_body": { "type": "string" },
		"http_output": { "$ref": "#/$defs/HTTPOutput" }
	},
	"allOf": [
        {
            "if": {
                "properties": {
                    "action": { "const": "create" }
                }
            },
            "then": {
                "required": ["input"]
            }
        },
				{
					"if": {
							"properties": {
									"action": { "const": "input" }
							}
					},
					"then": {
							"required": ["input"]
					}
				},
        {
            "if": {
                "properties": {
                    "action": { "const": "oauth_redirect" }
                }
            },
            "then": {
                "required": ["to", "redirect_uri"]
            }
        },
				{
				  "if": {
							"properties": {
									"action": { "const": "generate_totp_code" }
							}
					},
					"then": {
							"required": ["totp_secret"]
					}
				},
				{
				  "if": {
							"properties": {
									"action": { "const": "query" }
							}
					},
					"then": {
							"required": ["query"]
					}
				},
				{
				  "if": {
							"properties": {
									"action": { "const": "saml_request" }
							}
					},
					"then": {
							"required": [
								"saml_request_destination",
								"saml_request_binding"
							]
					}
				},
				{
				  "if": {
							"properties": {
									"action": { "const": "http_request" }
							}
					},
					"then": {
							"required": [
								"http_request_method",
								"http_request_url"
							]
					}
				}
    ]
}
`)

type Step struct {
	Name   string     `json:"name"`
	Action StepAction `json:"action"`

	// `action` == "create" or "input"
	Input string `json:"input"`

	// `action` == "oauth_redirect"
	To          string `json:"to"`
	RedirectURI string `json:"redirect_uri"`

	// `action` == "generate_totp_code"
	TOTPSecret string `json:"totp_secret"`

	// `action` == "input"
	Output *Output `json:"output"`

	// `action` == "query"
	Query       string       `json:"query"`
	QueryOutput *QueryOutput `json:"query_output"`

	// `action` == "saml_request"
	SAMLRequest            string                `json:"saml_request"`
	SAMLRequestDestination string                `json:"saml_request_destination"`
	SAMLRequestBinding     e2eclient.SAMLBinding `json:"saml_request_binding"`
	SAMLOutput             *SAMLOutput           `json:"saml_output"`

	// `action` == "http_request"
	HTTPRequestMethod  string            `json:"http_request_method"`
	HTTPRequestURL     string            `json:"http_request_url"`
	HTTPRequestHeaders map[string]string `json:"http_request_headers"`
	HTTPRequestBody    string            `json:"http_request_body"`
	HTTPOutput         *HTTPOutput       `json:"http_output"`
}

type StepAction string

const (
	StepActionCreate           StepAction = "create"
	StepActionInput            StepAction = "input"
	StepActionOAuthRedirect    StepAction = "oauth_redirect"
	StepActionGenerateTOTPCode StepAction = "generate_totp_code"
	StepActionQuery            StepAction = "query"
	StepActionSAMLRequest      StepAction = "saml_request"
	StepActionHTTPRequest      StepAction = "http_request"
)

var _ = TestCaseSchema.Add("SAMLOutput", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"http_status": { "type": "integer" },
		"redirect_path": { "type": "string" },
		"saml_response": { "$ref": "#/$defs/OuputSAMLResponse" }
	}
}
`)

type SAMLOutput struct {
	HTTPStatus   *float64           `json:"http_status"`
	RedirectPath *string            `json:"redirect_path"`
	SAMLResponse *OuputSAMLResponse `json:"saml_response"`
}

var _ = TestCaseSchema.Add("HTTPOutput", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"http_status": { "type": "integer" },
		"redirect_path": { "type": "string" },
		"saml_response": { "$ref": "#/$defs/OuputSAMLResponse" }
	}
}
`)

type HTTPOutput struct {
	HTTPStatus   *float64           `json:"http_status"`
	RedirectPath *string            `json:"redirect_path"`
	SAMLResponse *OuputSAMLResponse `json:"saml_response"`
}

var _ = TestCaseSchema.Add("OuputSAMLResponse", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"binding": { "$ref": "#/$defs/SAMLBinding" },
		"match": { "type": "string" }
	}
}
`)

type OuputSAMLResponse struct {
	Binding e2eclient.SAMLBinding `json:"binding"`
	Match   string                `json:"match"`
}

var _ = TestCaseSchema.Add("QueryOutput", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"rows": { "type": "string" }
	}
}
`)

type QueryOutput struct {
	Rows string `json:"rows"`
}

var _ = TestCaseSchema.Add("Output", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"result": { "type": "string" },
		"error": { "type": "string" }
	}
}
`)

type Output struct {
	Result string `json:"result"`
	Error  string `json:"error"`
}

var _ = TestCaseSchema.Add("StepResult", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"result": { "type": "string" },
		"error": { "type": "string" }
	}
}
`)

type StepResult struct {
	Result interface{} `json:"result"`
	Error  error       `json:"error"`
}

type ResultHTTPResponse struct {
	HTTPResponseHeaders map[string]string `json:"http_response_headers"`
}

func NewResultHTTPResponse(r *http.Response) *ResultHTTPResponse {
	headers := map[string]string{}
	for key, _ := range r.Header {
		headers[strings.ToLower(key)] = r.Header.Get(key)
	}
	return &ResultHTTPResponse{
		HTTPResponseHeaders: headers,
	}
}
