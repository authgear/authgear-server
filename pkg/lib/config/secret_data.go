package config

import (
	"encoding/json"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/authgear/authgear-server/pkg/util/jwkutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

var _ = SecretConfigSchema.Add("DatabaseCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"database_url": { "type": "string" },
		"database_schema": { "type": "string" }
	},
	"required": ["database_url"]
}
`)

type DatabaseCredentials struct {
	DatabaseURL    string `json:"database_url,omitempty"`
	DatabaseSchema string `json:"database_schema,omitempty"`
}

func (c *DatabaseCredentials) SensitiveStrings() []string {
	return []string{
		c.DatabaseURL,
	}
}

func (c *DatabaseCredentials) SetDefaults() {
	if c.DatabaseSchema == "" {
		c.DatabaseSchema = "public"
	}
}

var _ = SecretConfigSchema.Add("AuditDatabaseCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"database_url": { "type": "string" },
		"database_schema": { "type": "string" }
	},
	"required": ["database_url"]
}
`)

type AuditDatabaseCredentials struct {
	DatabaseURL    string `json:"database_url,omitempty"`
	DatabaseSchema string `json:"database_schema,omitempty"`
}

func (c *AuditDatabaseCredentials) SensitiveStrings() []string {
	return []string{
		c.DatabaseURL,
	}
}

func (c *AuditDatabaseCredentials) SetDefaults() {
	if c.DatabaseSchema == "" {
		c.DatabaseSchema = "public"
	}
}

var _ = SecretConfigSchema.Add("ElasticsearchCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"elasticsearch_url": { "type": "string" }
	},
	"required": ["elasticsearch_url"]
}
`)

type ElasticsearchCredentials struct {
	ElasticsearchURL string `json:"elasticsearch_url,omitempty"`
}

func (c *ElasticsearchCredentials) SensitiveStrings() []string {
	return []string{c.ElasticsearchURL}
}

var _ = SecretConfigSchema.Add("SearchDatabaseCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"database_url": { "type": "string" },
		"database_schema": { "type": "string" }
	},
	"required": ["database_url"]
}
`)

type SearchDatabaseCredentials struct {
	DatabaseURL    string `json:"database_url,omitempty"`
	DatabaseSchema string `json:"database_schema,omitempty"`
}

func (c *SearchDatabaseCredentials) SensitiveStrings() []string {
	return []string{c.DatabaseURL}
}

var _ = SecretConfigSchema.Add("RedisCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"redis_url": { "type": "string" }
	},
	"required": ["redis_url"]
}
`)

type RedisCredentials struct {
	RedisURL string `json:"redis_url,omitempty"`
}

func (c *RedisCredentials) SensitiveStrings() []string {
	return []string{c.RedisURL}
}

var _ = SecretConfigSchema.Add("AnalyticRedisCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"redis_url": { "type": "string" }
	},
	"required": ["redis_url"]
}
`)

type AnalyticRedisCredentials struct {
	RedisURL string `json:"redis_url,omitempty"`
}

func (c *AnalyticRedisCredentials) SensitiveStrings() []string {
	return []string{c.RedisURL}
}

var _ = SecretConfigSchema.Add("OAuthSSOProviderCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"items": {
			"type": "array",
			"items": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"alias": {
						"type": "string"
					},
					"client_secret": {
						"type": "string"
					}
				},
				"required": ["alias", "client_secret"]
			}
		}
	},
	"required": ["items"]
}
`)

type OAuthSSOProviderCredentials struct {
	Items []OAuthSSOProviderCredentialsItem `json:"items,omitempty"`
}

func (c *OAuthSSOProviderCredentials) Lookup(alias string) (*OAuthSSOProviderCredentialsItem, bool) {
	for _, item := range c.Items {
		if item.Alias == alias {
			ii := item
			return &ii, true
		}
	}
	return nil, false
}

func (c *OAuthSSOProviderCredentials) SensitiveStrings() []string {
	var out []string
	for _, item := range c.Items {
		out = append(out, item.SensitiveStrings()...)
	}
	return out
}

type OAuthSSOProviderCredentialsItem struct {
	Alias string `json:"alias,omitempty"`
	// client_secret can be empty string for inactive providers
	ClientSecret string `json:"client_secret"`
}

func (c *OAuthSSOProviderCredentialsItem) SensitiveStrings() []string {
	return []string{c.ClientSecret}
}

var _ = SecretConfigSchema.Add("SMTPMode", `
{
	"type": "string",
	"enum": ["normal", "ssl"]
}
`)

type SMTPMode string

const (
	SMTPModeNormal SMTPMode = "normal"
	SMTPModeSSL    SMTPMode = "ssl"
)

var _ = SecretConfigSchema.Add("SMTPServerCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"host": { "type": "string" },
		"port": { "type": "integer", "minimum": 1, "maximum": 65535 },
		"mode": { "$ref": "#/$defs/SMTPMode" },
		"username": { "type": "string" },
		"password": { "type": "string" },
		"sender": { "type": "string", "format": "email-name-addr" }
	},
	"required": ["host", "port", "username", "password"]
}
`)

type SMTPServerCredentialsSecretItem SecretItem

func (s *SMTPServerCredentialsSecretItem) GetData() *SMTPServerCredentials {
	return s.Data.(*SMTPServerCredentials)
}

type SMTPServerCredentials struct {
	Host     string   `json:"host,omitempty"`
	Port     int      `json:"port,omitempty"`
	Mode     SMTPMode `json:"mode,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Sender   string   `json:"sender,omitempty"`
}

func (c *SMTPServerCredentials) SensitiveStrings() []string {
	return []string{
		c.Host,
		c.Username,
		c.Password,
	}
}

func (c *SMTPServerCredentials) SetDefaults() {
	if c.Mode == "" {
		c.Mode = SMTPModeNormal
	}
}

var _ = SecretConfigSchema.Add("TwilioCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"credential_type": {
			"type": "string",
			"enum": ["api_key", "auth_token"]
		},
		"account_sid": { "type": "string" },
		"auth_token": { "type": "string" },
		"api_key_sid": { "type": "string" },
		"api_key_secret": { "type": "string" },
		"message_service_sid": { "type": "string" },
		"from": { "type": "string" }
	},
	"required": ["account_sid"],
	"allOf": [
		{
			"if": {
				"properties": { "credential_type": { "const": "api_key" } },
				"required": ["credential_type"]
			},
			"then": { "required": ["api_key_sid", "api_key_secret"] },
			"else": { "required": ["auth_token"] }
		}
	]
}
`)

type TwilioCredentialType string

const (
	TwilioCredentialTypeAPIKey    TwilioCredentialType = "api_key"
	TwilioCredentialTypeAuthToken TwilioCredentialType = "auth_token"
)

type TwilioCredentials struct {
	// For write only, always use GetCredentialType to read the value
	CredentialType_WriteOnly *TwilioCredentialType `json:"credential_type,omitempty"`
	AccountSID               string                `json:"account_sid,omitempty"`
	AuthToken                string                `json:"auth_token,omitempty"`
	APIKeySID                string                `json:"api_key_sid,omitempty"`
	APIKeySecret             string                `json:"api_key_secret,omitempty"`
	MessagingServiceSID      string                `json:"message_service_sid,omitempty"`
	From                     string                `json:"from,omitempty"`
}

func (c *TwilioCredentials) SensitiveStrings() []string {
	return []string{
		c.AccountSID,
		c.AuthToken,
		c.APIKeySID,
		c.APIKeySecret,
		c.MessagingServiceSID,
	}
}

func (c *TwilioCredentials) GetCredentialType() TwilioCredentialType {
	if c.CredentialType_WriteOnly == nil {
		// Old data does not specify credential_type. Treat it as auth_token.
		return TwilioCredentialTypeAuthToken
	}
	return *c.CredentialType_WriteOnly
}

var _ = SecretConfigSchema.Add("NexmoCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"api_key": { "type": "string" },
		"api_secret": { "type": "string" }
	},
	"required": ["api_key", "api_secret"]
}
`)

type NexmoCredentials struct {
	APIKey    string `json:"api_key,omitempty"`
	APISecret string `json:"api_secret,omitempty"`
}

func (c *NexmoCredentials) SensitiveStrings() []string {
	return []string{
		c.APIKey,
		c.APISecret,
	}
}

var _ = SecretConfigSchema.Add("JWS", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"keys": {
			"type": "array",
			"items": { "$ref": "#/$defs/JWK" },
			"minItems": 1
		}
	},
	"required": ["keys"]
}
`)

var _ = SecretConfigSchema.Add("OAuthKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type OAuthKeyMaterials struct {
	jwk.Set
}

var _ json.Marshaler = &OAuthKeyMaterials{}
var _ json.Unmarshaler = &OAuthKeyMaterials{}

func (c *OAuthKeyMaterials) MarshalJSON() ([]byte, error) {
	return c.Set.(interface{}).(json.Marshaler).MarshalJSON()
}
func (c *OAuthKeyMaterials) UnmarshalJSON(b []byte) error {
	if c.Set == nil {
		c.Set = jwk.NewSet()
	}
	return c.Set.(interface{}).(json.Unmarshaler).UnmarshalJSON(b)
}

func (c *OAuthKeyMaterials) SensitiveStrings() []string {
	return nil
}

var _ = SecretConfigSchema.Add("CSRFKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type CSRFKeyMaterials struct {
	jwk.Set
}

var _ json.Marshaler = &CSRFKeyMaterials{}
var _ json.Unmarshaler = &CSRFKeyMaterials{}

func (c *CSRFKeyMaterials) MarshalJSON() ([]byte, error) {
	return c.Set.(interface{}).(json.Marshaler).MarshalJSON()
}
func (c *CSRFKeyMaterials) UnmarshalJSON(b []byte) error {
	if c.Set == nil {
		c.Set = jwk.NewSet()
	}
	return c.Set.(interface{}).(json.Unmarshaler).UnmarshalJSON(b)
}

func (c *CSRFKeyMaterials) SensitiveStrings() []string {
	keys, _ := jwkutil.ExtractOctetKeys(c.Set)
	return slice.ToStringSlice(keys)
}

var _ = SecretConfigSchema.Add("WebhookKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type WebhookKeyMaterials struct {
	jwk.Set
}

var _ json.Marshaler = &WebhookKeyMaterials{}
var _ json.Unmarshaler = &WebhookKeyMaterials{}

func (c *WebhookKeyMaterials) MarshalJSON() ([]byte, error) {
	return c.Set.(interface{}).(json.Marshaler).MarshalJSON()
}
func (c *WebhookKeyMaterials) UnmarshalJSON(b []byte) error {
	if c.Set == nil {
		c.Set = jwk.NewSet()
	}
	return c.Set.(interface{}).(json.Unmarshaler).UnmarshalJSON(b)
}

func (c *WebhookKeyMaterials) SensitiveStrings() []string {
	keys, _ := jwkutil.ExtractOctetKeys(c.Set)
	return slice.ToStringSlice(keys)
}

var _ = SecretConfigSchema.Add("AdminAPIAuthKey", `{ "$ref": "#/$defs/JWS" }`)

type AdminAPIAuthKey struct {
	jwk.Set
}

var _ json.Marshaler = &AdminAPIAuthKey{}
var _ json.Unmarshaler = &AdminAPIAuthKey{}

func (c *AdminAPIAuthKey) MarshalJSON() ([]byte, error) {
	return c.Set.(interface{}).(json.Marshaler).MarshalJSON()
}
func (c *AdminAPIAuthKey) UnmarshalJSON(b []byte) error {
	if c.Set == nil {
		c.Set = jwk.NewSet()
	}
	return c.Set.(interface{}).(json.Unmarshaler).UnmarshalJSON(b)
}

func (c *AdminAPIAuthKey) SensitiveStrings() []string {
	return nil
}

var _ = SecretConfigSchema.Add("ImagesKeyMaterials", `{ "$ref": "#/$defs/JWS" }`)

type ImagesKeyMaterials struct {
	jwk.Set
}

var _ json.Marshaler = &ImagesKeyMaterials{}
var _ json.Unmarshaler = &ImagesKeyMaterials{}

func (c *ImagesKeyMaterials) MarshalJSON() ([]byte, error) {
	return c.Set.(interface{}).(json.Marshaler).MarshalJSON()
}
func (c *ImagesKeyMaterials) UnmarshalJSON(b []byte) error {
	if c.Set == nil {
		c.Set = jwk.NewSet()
	}
	return c.Set.(interface{}).(json.Unmarshaler).UnmarshalJSON(b)
}

func (c *ImagesKeyMaterials) SensitiveStrings() []string {
	keys, _ := jwkutil.ExtractOctetKeys(c.Set)
	return slice.ToStringSlice(keys)
}

var _ = SecretConfigSchema.Add("WATICredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"whatsapp_phone_number": { "type": "string" },
		"webhook_auth": { "type": "string" }
	},
	"required": ["whatsapp_phone_number", "webhook_auth"]
}
`)

// WATICredentials is deprecated, don't use it
type WATICredentials struct {
	WhatsappPhoneNumber string `json:"whatsapp_phone_number,omitempty"`
	WebhookAuth         string `json:"webhook_auth,omitempty"`
}

func (c *WATICredentials) SensitiveStrings() []string {
	return []string{
		c.WhatsappPhoneNumber,
		c.WebhookAuth,
	}
}

var _ = SecretConfigSchema.Add("OAuthClientCredentials", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"items": {
			"type": "array",
			"items": {
				"type": "object",
				"additionalProperties": false,
				"properties": {
					"client_id": {
						"type": "string"
					},
					"keys": {
						"type": "array",
						"items": { "$ref": "#/$defs/JWK" },
						"minItems": 1
					}
				},
				"required": ["client_id", "keys"]
			}
		}
	},
	"required": ["items"]
}
`)

type OAuthClientCredentials struct {
	Items []OAuthClientCredentialsItem `json:"items,omitempty"`
}

func (c *OAuthClientCredentials) Lookup(clientID string) (*OAuthClientCredentialsItem, bool) {
	for _, item := range c.Items {
		if item.ClientID == clientID {
			ii := item
			return &ii, true
		}
	}
	return nil, false
}

func (c *OAuthClientCredentials) SensitiveStrings() []string {
	var out []string
	for _, item := range c.Items {
		out = append(out, item.SensitiveStrings()...)
	}
	return out
}

type OAuthClientCredentialsItem struct {
	// It is important to update `MarshalJSON` and `UnmarshalJSON` functions
	// when updating fields of OAuthClientCredentialsItem
	ClientID string `json:"client_id,omitempty"`
	OAuthClientCredentialsKeySet
}

func (c *OAuthClientCredentialsItem) MarshalJSON() ([]byte, error) {
	// convert key to json string
	keyJSON, err := json.Marshal(c.Set)
	if err != nil {
		return nil, err
	}

	// convert the json string to map
	var dataMap map[string]interface{}
	err = json.Unmarshal(keyJSON, &dataMap)
	if err != nil {
		return nil, err
	}

	// add fields to the map
	dataMap["client_id"] = c.ClientID

	// covert map to json
	return json.Marshal(dataMap)
}

func (c *OAuthClientCredentialsItem) UnmarshalJSON(b []byte) error {
	// unmarshal the keys
	err := json.Unmarshal(b, &c.OAuthClientCredentialsKeySet)
	if err != nil {
		return err
	}

	// unmarshal to the map
	var dataMap map[string]interface{}
	err = json.Unmarshal(b, &dataMap)
	if err != nil {
		return err
	}

	// fill the fields
	clientID, ok := dataMap["client_id"].(string)
	if ok {
		c.ClientID = clientID
	}

	return nil
}

type OAuthClientCredentialsKeySet struct {
	jwk.Set
}

var _ json.Marshaler = &OAuthClientCredentialsKeySet{}
var _ json.Unmarshaler = &OAuthClientCredentialsKeySet{}

func (c *OAuthClientCredentialsKeySet) MarshalJSON() ([]byte, error) {
	return c.Set.(interface{}).(json.Marshaler).MarshalJSON()
}
func (c *OAuthClientCredentialsKeySet) UnmarshalJSON(b []byte) error {
	if c.Set == nil {
		c.Set = jwk.NewSet()
	}
	return c.Set.(interface{}).(json.Unmarshaler).UnmarshalJSON(b)
}

func (c *OAuthClientCredentialsKeySet) SensitiveStrings() []string {
	keys, _ := jwkutil.ExtractOctetKeys(c.Set)
	return slice.ToStringSlice(keys)
}
