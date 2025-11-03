package testrunner

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/beevik/etree"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/pkg/lib/authn/challenge"
)

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
	BeforeHookTypeUserImport      BeforeHookType = "user_import"
	BeforeHookTypeCustomSQL       BeforeHookType = "custom_sql"
	BeforeHookTypeCreateSession   BeforeHookType = "create_session"
	BeforeHookTypeCreateChallenge BeforeHookType = "create_challenge"
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

var _ = TestCaseSchema.Add("BeforeHookCreateSession", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"session_type": { "type": "string", "enum": ["idp", "offline_grant"], "description": "Session Type" },
		"session_id": { "type": "string", "description": "Session ID" },
		"client_id": { "type": "string", "description": "Client ID" },
		"token": { "type": "string", "description": "Token to access the session" },
		"access_token": { "type": "string", "description": "Access Token to access the offline_grant session" },
		"select_user_id_sql":  { "type": "string", "description": "SQL to select an user id for the session" }
	},
	"required": ["session_type", "session_id", "token", "select_user_id_sql"],
	"allOf": [
		{
			"if": {
				"properties": {
					"session_type": { "const": "offline_grant" }
				}
			},
			"then": {
				"required": ["client_id"]
			}
		}
	]
}
`)

type BeforeHookCreateSession struct {
	SessionType     string `json:"session_type"`
	SessionID       string `json:"session_id"`
	ClientID        string `json:"client_id"`
	Token           string `json:"token"`
	AccessToken     string `json:"access_token"`
	SelectUserIDSQL string `json:"select_user_id_sql"`
}

var _ = TestCaseSchema.Add("BeforeHookCreateChallenge", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"token": { "type": "string" },
		"purpose": {
			"type": "string",
			"enum": [
				"anonymous_request",
				"biometric_request",
				"app2app_request"
			]
		}
	},
	"required": ["token", "purpose"]
}
`)

type BeforeHookCreateChallenge struct {
	Token   string            `json:"token"`
	Purpose challenge.Purpose `json:"purpose"`
}

var _ = TestCaseSchema.Add("BeforeHook", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"type": { "type": "string", "enum": ["user_import", "custom_sql", "create_session", "create_challenge"] },
		"user_import": { "type": "string" },
		"custom_sql": { "$ref": "#/$defs/BeforeHookCustomSQL" },
		"create_session": { "$ref": "#/$defs/BeforeHookCreateSession" },
		"create_challenge": { "$ref": "#/$defs/BeforeHookCreateChallenge" }
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
			},
			{
				"if": { "properties": { "type": { "const": "create_session" } } },
				"then": {
					"required": ["create_session"]
				}
			},
			{
				"if": { "properties": { "type": { "const": "create_challenge" } } },
				"then": {
					"required": ["create_challenge"]
				}
			}
		]
}
`)

type BeforeHook struct {
	Type            BeforeHookType             `json:"type"`
	UserImport      string                     `json:"user_import"`
	CustomSQL       *BeforeHookCustomSQL       `json:"custom_sql"`
	CreateSession   *BeforeHookCreateSession   `json:"create_session"`
	CreateChallenge *BeforeHookCreateChallenge `json:"create_challenge"`
}

var _ = TestCaseSchema.Add("SAMLBinding", `
{
	"type": "string",
	"enum": [
		"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect",
		"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
	]
}
`)

type AdminAPIRequest struct {
	Query     string `json:"query"`
	Variables string `json:"variables"`
}

var _ = TestCaseSchema.Add("AdminAPIRequest", `
{
    "type": "object",
    "properties": {
        "query": { "type": "string" },
        "variables": { "type": "string" }
    },
    "required": ["query", "variables"]
}
`)

var _ = TestCaseSchema.Add("AdminAPIOutput", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"result": { "type": "string" }
	}
}
`)

type AdminAPIUserImportRequest struct {
	JSONDocument string `json:"json_document"`
}

var _ = TestCaseSchema.Add("AdminAPIUserImportRequest", `
{
    "type": "object",
    "properties": {
        "json_document": { "type": "string" }
    },
    "required": ["json_document"]
}
`)

var _ = TestCaseSchema.Add("AdminAPIUserImportOutput", `
{
	"type": "object",
	"properties": {
		"result": { "type": "string" },
		"error": { "type": "string" }
	}
}
`)

var _ = TestCaseSchema.Add("Step", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string" },
		"action": { "type": "string", "enum": [
			"sleep",
			"create",
			"input",
			"oauth_redirect",
			"generate_totp_code",
			"query",
			"saml_request",
			"http_request",
			"oauth_setup",
			"oauth_exchange_code",
			"admin_api_graphql",
			"admin_api_user_import_create",
			"admin_api_user_import_get"
		]},
		"sleep_for": { "type": "string", "format": "x_duration_string" },
		"input": { "type": "string" },
		"to": { "type": "string" },
		"redirect_uri": { "type": "string" },
		"totp_secret": { "type": "string" },
		"output": { "$ref": "#/$defs/Output" },
		"query": { "type": "string" },
		"query_output": { "$ref": "#/$defs/QueryOutput" },
		"saml_output": { "$ref": "#/$defs/SAMLOutput" },
		"saml_element": { "type": "string" },
		"saml_element_name": { "type": "string", "enum": ["SAMLRequest", "SAMLResponse"] },
		"saml_request_destination": { "type": "string" },
		"saml_request_binding": { "$ref": "#/$defs/SAMLBinding" },
		"saml_request_relay_state": { "type": "string" },
		"saml_request_session_cookie": { "$ref": "#/$defs/SessionCookie" },
		"http_request_method": { "type": "string" },
		"http_request_url": { "type": "string" },
		"http_request_headers": { "type": "object" },
		"http_request_body": { "type": "string" },
		"http_request_form_urlencoded_body": { "type": "object" },
		"http_request_session_cookie": { "$ref": "#/$defs/SessionCookie" },
		"http_output": { "$ref": "#/$defs/HTTPOutput" },
		"oauth_exchange_code_code_verifier": { "type": "string" },
		"oauth_exchange_code_redirect_uri": { "type": "string" },
		"admin_api_request": { "$ref": "#/$defs/AdminAPIRequest" },
		"admin_api_output": { "$ref": "#/$defs/AdminAPIOutput" },
		"admin_api_user_import_request": { "$ref": "#/$defs/AdminAPIUserImportRequest" },
		"admin_api_user_import_output": { "$ref": "#/$defs/AdminAPIUserImportOutput" },
		"admin_api_user_import_id": { "type": "string" }
	},
	"allOf": [
		{
			"if": {
				"properties": {
					"action": { "const": "sleep" }
				}
			},
			"then": {
				"required": ["sleep_for"]
			}
		},
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
								"saml_element",
								"saml_element_name",
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
				},
				{
					"if": {
						"properties": {
							"action": { "const": "oauth_exchange_code" }
						}
					},
					"then": {
						"required": [
							"oauth_exchange_code_code_verifier",
							"oauth_exchange_code_redirect_uri"
						]
					}
				},
				{
					"if": {
						"properties": {
							"action": { "const": "admin_api_graphql" }
						}
					},
					"then": {
						"required": [
							"admin_api_request"
						]
					}
				},
				{
					"if": {
						"properties": {
							"action": { "const": "admin_api_user_import_create" }
						}
					},
					"then": {
						"required": [
							"admin_api_user_import_request"
						]
					}
				},
				{
					"if": {
						"properties": {
							"action": { "const": "admin_api_user_import_get" }
						}
					},
					"then": {
						"required": [
							"admin_api_user_import_id",
							"admin_api_user_import_output"
						]
					}
				}
    ]
}
`)

type Step struct {
	Name   string     `json:"name"`
	Action StepAction `json:"action"`

	// `action` == "sleep"
	SleepFor string `json:"sleep_for,omitzero"`

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
	SAMLElement              string                `json:"saml_element"`
	SAMLElementName          string                `json:"saml_element_name"`
	SAMLRequestDestination   string                `json:"saml_request_destination"`
	SAMLRequestBinding       e2eclient.SAMLBinding `json:"saml_request_binding"`
	SAMLRequestRelayState    string                `json:"saml_request_relay_state"`
	SAMLRequestSessionCookie *SessionCookie        `json:"saml_request_session_cookie"`
	SAMLOutput               *SAMLOutput           `json:"saml_output"`

	// `action` == "http_request"
	HTTPRequestMethod             string            `json:"http_request_method"`
	HTTPRequestURL                string            `json:"http_request_url"`
	HTTPRequestHeaders            map[string]string `json:"http_request_headers"`
	HTTPRequestBody               string            `json:"http_request_body"`
	HTTPRequestFormURLEncodedBody map[string]string `json:"http_request_form_urlencoded_body"`
	HTTPRequestSessionCookie      *SessionCookie    `json:"http_request_session_cookie"`
	HTTPOutput                    *HTTPOutput       `json:"http_output"`

	// `action` == "oauth_exchange_code"
	OAuthExchangeCodeCodeVerifier string `json:"oauth_exchange_code_code_verifier"`
	OAuthExchangeCodeRedirectURI  string `json:"oauth_exchange_code_redirect_uri"`

	// `action` == "admin_api_graphql"
	AdminAPIRequest *AdminAPIRequest `json:"admin_api_request"`
	AdminAPIOutput  *AdminAPIOutput  `json:"admin_api_output"`

	// `action` == "admin_api_user_import_create"
	AdminAPIUserImportRequest *AdminAPIUserImportRequest `json:"admin_api_user_import_request"`
	// `action` == "admin_api_user_import_get"
	AdminAPIUserImportID string `json:"admin_api_user_import_id"`
	// `action` == "admin_api_user_import_create" or "admin_api_user_import_get"
	AdminAPIUserImportOutput *AdminAPIUserImportOutput `json:"admin_api_user_import_output"`
}

type StepAction string

const (
	StepActionSleep                    StepAction = "sleep"
	StepActionCreate                   StepAction = "create"
	StepActionInput                    StepAction = "input"
	StepActionOAuthRedirect            StepAction = "oauth_redirect"
	StepActionGenerateTOTPCode         StepAction = "generate_totp_code"
	StepActionQuery                    StepAction = "query"
	StepActionSAMLRequest              StepAction = "saml_request"
	StepActionHTTPRequest              StepAction = "http_request"
	StepActionOAuthSetup               StepAction = "oauth_setup"
	StepActionOAuthExchangeCode        StepAction = "oauth_exchange_code"
	StepActionAdminAPIQuery            StepAction = "admin_api_graphql"
	StepActionAdminAPIUserImportCreate StepAction = "admin_api_user_import_create"
	StepActionAdminAPIUserImportGet    StepAction = "admin_api_user_import_get"
)

var _ = TestCaseSchema.Add("SessionCookie", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"idp_session_id": { "type": "string" },
		"idp_session_token": { "type": "string" }
	}
}
`)

type SessionCookie struct {
	IDPSessionID    string `json:"idp_session_id"`
	IDPSessionToken string `json:"idp_session_token"`
}

var _ = TestCaseSchema.Add("SAMLOutput", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"http_status": { "type": "integer" },
		"redirect_path": { "type": "string" },
		"saml_element": { "$ref": "#/$defs/OuputSAMLElement" }
	}
}
`)

type SAMLOutput struct {
	HTTPStatus   *float64          `json:"http_status"`
	RedirectPath *string           `json:"redirect_path"`
	SAMLElement  *OuputSAMLElement `json:"saml_element"`
}

var _ = TestCaseSchema.Add("HTTPOutput", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"http_status": { "type": "integer" },
		"redirect_path": { "type": "string" },
		"saml_element": { "$ref": "#/$defs/OuputSAMLElement" },
		"json_body": { "type": "string" }
	}
}
`)

type HTTPOutput struct {
	HTTPStatus   *float64          `json:"http_status"`
	RedirectPath *string           `json:"redirect_path"`
	SAMLElement  *OuputSAMLElement `json:"saml_element"`
	JSONBody     *string           `json:"json_body"`
}

var _ = TestCaseSchema.Add("OuputSAMLElement", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"element_name": { "type": "string" },
		"binding": { "$ref": "#/$defs/SAMLBinding" },
		"match": { "type": "string" }
	},
	"required": ["element_name", "binding", "match"]
}
`)

type OuputSAMLElement struct {
	ElementName string                `json:"element_name"`
	Binding     e2eclient.SAMLBinding `json:"binding"`
	Match       string                `json:"match"`
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

// AdminAPIOutput is used for admin_api_graphql output validation
type AdminAPIOutput struct {
	Result string `json:"result"`
}

type AdminAPIUserImportOutput struct {
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
	Step   *Step       `json:"step"`
	Result interface{} `json:"result"`
	Error  error       `json:"error"`
}

type ResultHTTPResponse struct {
	HTTPResponseHeaders map[string]string `json:"http_response_headers"`
	SAMLRelayState      string            `json:"saml_relay_state"`
}

func NewResultHTTPResponse(r *http.Response) *ResultHTTPResponse {
	headers := map[string]string{}
	for key, _ := range r.Header {
		headers[strings.ToLower(key)] = r.Header.Get(key)
	}

	samlRelayState := extractRelayState(r)

	return &ResultHTTPResponse{
		HTTPResponseHeaders: headers,
		SAMLRelayState:      samlRelayState,
	}
}

func extractRelayState(r *http.Response) string {

	// Try to read relay state from redirect location or body
	if location := r.Header.Get("location"); location != "" {
		locationURL, err := url.Parse(location)
		if err == nil {
			relayState := locationURL.Query().Get("RelayState")
			if relayState != "" {
				return relayState
			}
		}
	}

	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			// Allow later code to read the body again
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			doc := etree.NewDocument()
			err = doc.ReadFromString(string(bodyBytes))
			if err != nil {
				return ""
			}
			samlRelayStateEl := doc.FindElement("./html/body/form/input[@name='RelayState']")
			if samlRelayStateEl != nil {
				attrEl := samlRelayStateEl.SelectAttr("value")
				if attrEl != nil {
					return attrEl.Value
				}
			}
		}
	}

	return ""
}
