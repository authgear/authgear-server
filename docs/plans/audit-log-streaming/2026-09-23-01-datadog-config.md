# Audit Log Streaming, type datadog — Part 01: Config

Implements the configuration surface of `type: datadog` and `transport: http`, from [`docs/specs/audit-log-streaming.md`](../../specs/audit-log-streaming.md).

This is an increment on the shipped syslog feature, whose plan is [2026-09-18-01-config](./2026-09-18-01-config.md), [-02-runtime](./2026-09-18-02-runtime.md) and [-03-e2e](./2026-09-18-03-e2e.md). Those three describe the code as it exists today; this part and its two siblings describe only what changes.

## Goal / scope

After this part:

- `type: datadog` and `transport: http` parse, default and validate in `authgear.yaml`, including the encoding × transport matrix and the `datadog.site` / `http.endpoint` mutual exclusion.
- `telemetry.audit_logs.streams.datadog` parses and validates in `authgear.secrets.yaml`, and a `type: datadog` stream without its `api_key` is a configuration error.
- An orphaned datadog secret item is pruned when `authgear.yaml` is next saved, as a tls item already is.
- Nothing is delivered to Datadog. A configured datadog stream is silently skipped by the sender until part 02.

Part 02 adds the log object, the chunked gzipped HTTP delivery and the wiring. Part 03 adds e2e coverage.

## 1. Config model and schema

### 1.1 New enum values — `pkg/lib/config/telemetry.go`

```go
const (
	TelemetryAuditLogStreamTypeSyslog  TelemetryAuditLogStreamType = "syslog"
	TelemetryAuditLogStreamTypeDatadog TelemetryAuditLogStreamType = "datadog"
)

const (
	TelemetryAuditLogStreamTransportTCP  TelemetryAuditLogStreamTransport = "tcp"
	TelemetryAuditLogStreamTransportHTTP TelemetryAuditLogStreamTransport = "http"
)
```

### 1.2 The datadog object — `pkg/lib/config/telemetry.go`

```go
// DatadogSite is the site parameter of a Datadog organization, not the
// display name of one. It selects the intake endpoint; see
// LogsIntakeURL.
//
// It is a free-form string, not a closed enum (see §1.5): which sites
// exist is Datadog's deployment to govern, not this project's.
// DatadogSiteUS1 is the only named constant, because it is the only site
// this package needs to refer to -- the default. A second named constant
// would just be an unenforced, staleness-prone copy of Datadog's site
// list, the exact thing the relaxation was meant to avoid.
type DatadogSite string

const (
	DatadogSiteUS1 DatadogSite = "datadoghq.com"
)

// LogsIntakeURL returns the logs HTTP intake URL of s. Every Datadog site
// spells this host the same way, so this is a single interpolation
// rather than a switch -- and it is why s does not need to be validated
// against a known list for this to work correctly.
func (s DatadogSite) LogsIntakeURL() string {
	return fmt.Sprintf("https://http-intake.logs.%s/api/v2/logs", string(s))
}

type TelemetryAuditLogStreamDatadogConfig struct {
	Site    DatadogSite       `json:"site,omitempty"`
	Service string            `json:"service,omitempty"`
	Source  string            `json:"source,omitempty"`
	Tags    map[string]string `json:"tags,omitempty"`
}

func (c *TelemetryAuditLogStreamDatadogConfig) SetDefaults() {
	if c.Site == "" {
		c.Site = DatadogSiteUS1
	}
	if c.Service == "" {
		c.Service = "authgear"
	}
	if c.Source == "" {
		c.Source = "authgear"
	}
}
```

`LogsIntakeURL` is a method on the config enum, next to it, for the same reason `SyslogFacility.Code()` is: it is the one place the mapping from a configured name to a protocol value is written, and it is called once per stream at resolve time (part 02 §3.1).

`Tags` is left nil by `SetDefaults`. The spec's default is `{}`, and a nil map ranges as an empty one, so the two are the same thing to the encoder; `omitempty` then keeps the effective config free of an empty object. This is the one place the plan deliberately does not materialise — see §1.4, where materialisation is the hazard being avoided.

### 1.3 The http object — `pkg/lib/config/telemetry.go`

```go
type TelemetryAuditLogStreamHTTPConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
}
```

No `SetDefaults`. The spec's default for `endpoint` is "derived from the encoding object", which is a runtime resolution (part 02 §3.1), not a config default: writing the derived URL into the config here would make `datadog.site` and `http.endpoint` both set in the effective config, which is exactly the state §1.5's schema rejects.

### 1.4 Per-stream object materialisation

`TelemetryAuditLogStreamConfig` gains two fields, and all four object fields gain `nullable:"true"`:

```go
type TelemetryAuditLogStreamConfig struct {
	Name      string                           `json:"name,omitempty"`
	Type      TelemetryAuditLogStreamType      `json:"type,omitempty"`
	Transport TelemetryAuditLogStreamTransport `json:"transport,omitempty"`

	// The four objects below are nullable so that SetFieldDefaults does not
	// materialize them. A stream carries exactly the encoding object its
	// type selects and exactly the transport object its transport selects;
	// SetDefaults below allocates those two and leaves the other two nil.
	TCP     *TelemetryAuditLogStreamTCPConfig     `json:"tcp,omitempty" nullable:"true"`
	HTTP    *TelemetryAuditLogStreamHTTPConfig    `json:"http,omitempty" nullable:"true"`
	Syslog  *TelemetryAuditLogStreamSyslogConfig  `json:"syslog,omitempty" nullable:"true"`
	Datadog *TelemetryAuditLogStreamDatadogConfig `json:"datadog,omitempty" nullable:"true"`
}

func (c *TelemetryAuditLogStreamConfig) SetDefaults() {
	switch c.Type {
	case TelemetryAuditLogStreamTypeSyslog:
		if c.Syslog == nil {
			c.Syslog = &TelemetryAuditLogStreamSyslogConfig{}
		}
		c.Syslog.SetDefaults()
	case TelemetryAuditLogStreamTypeDatadog:
		if c.Datadog == nil {
			c.Datadog = &TelemetryAuditLogStreamDatadogConfig{}
		}
		c.Datadog.SetDefaults()
	}

	switch c.Transport {
	case TelemetryAuditLogStreamTransportTCP:
		if c.TCP == nil {
			c.TCP = &TelemetryAuditLogStreamTCPConfig{}
		}
		c.TCP.SetDefaults()
	case TelemetryAuditLogStreamTransportHTTP:
		if c.HTTP == nil {
			c.HTTP = &TelemetryAuditLogStreamHTTPConfig{}
		}
	}
}

func (c *TelemetryAuditLogStreamTCPConfig) SetDefaults() {
	if c.TLS == nil {
		c.TLS = &TelemetryAuditLogStreamTCPTLSConfig{}
	}
}
```

This is the one non-obvious change in this part, and it is load-bearing, so the reasoning is written out.

**Why it is needed.** Without `nullable:"true"`, `SetFieldDefaults` (`pkg/lib/config/default.go:33-36`) instantiates every nil struct pointer it walks. A datadog stream would then come out of `Parse` carrying a defaulted `syslog` object (`facility: local0`, `app_name: authgear`, `structured_data_id: authgear`) and a `tcp` object, and a syslog stream would come out carrying a defaulted `datadog` object. §1.5's schema rejects exactly those combinations, so the parsed-then-re-marshalled config would no longer validate — the effective-config view (`AuthgearYAMLDescriptor.ViewResources`) would show objects the project never wrote, and a round trip through it would be rejected on save.

**Why it is safe.** `SetFieldDefaults`'s pointer case recurses into the pointee **before** calling `SetDefaults()` on the pointer (`default.go:40-47`), so a parent's `SetDefaults` runs after its children's and cannot be undone by the walk. Because the four fields are skipped by the walk, their own `SetDefaults` is no longer reached by it, which is why the parent calls them explicitly, and why `TelemetryAuditLogStreamTCPConfig` needs a `SetDefaults` of its own to materialise `TLS` (the walk used to do that).

**What it changes for an existing config.** Nothing. A `syslog` + `tcp` stream still ends up with both objects allocated and defaulted, by the same values, so `stream.Syslog.Facility`, `stream.TCP.Address` and `stream.TCP.TLS.Enabled` stay safe to read without nil guards in `resolveStream` (part 02 §3.1 keeps that contract, narrowed to the matching type/transport).

### 1.5 Stream schema — `pkg/lib/config/telemetry.go`

`TelemetryAuditLogStreamConfig`'s schema is replaced with:

```go
var _ = Schema.Add("TelemetryAuditLogStreamConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"name": { "type": "string", "pattern": "^[a-zA-Z0-9_-]{1,63}$" },
		"type": { "type": "string", "enum": ["syslog", "datadog"] },
		"transport": { "type": "string", "enum": ["tcp", "http"] },
		"tcp": { "$ref": "#/$defs/TelemetryAuditLogStreamTCPConfig" },
		"http": { "$ref": "#/$defs/TelemetryAuditLogStreamHTTPConfig" },
		"syslog": { "$ref": "#/$defs/TelemetryAuditLogStreamSyslogConfig" },
		"datadog": { "$ref": "#/$defs/TelemetryAuditLogStreamDatadogConfig" }
	},
	"required": ["name", "type", "transport"],
	"allOf": [
		{
			"if": { "properties": { "type": { "const": "syslog" } }, "required": ["type"] },
			"then": {
				"required": ["syslog"],
				"not": { "required": ["datadog"] },
				"properties": { "transport": { "enum": ["tcp"] } }
			}
		},
		{
			"if": { "properties": { "type": { "const": "datadog" } }, "required": ["type"] },
			"then": {
				"not": { "required": ["syslog"] },
				"properties": { "transport": { "enum": ["http"] } }
			}
		},
		{
			"if": { "properties": { "transport": { "const": "tcp" } }, "required": ["transport"] },
			"then": {
				"required": ["tcp"],
				"not": { "required": ["http"] }
			}
		},
		{
			"if": { "properties": { "transport": { "const": "http" } }, "required": ["transport"] },
			"then": { "not": { "required": ["tcp"] } }
		},
		{
			"if": {
				"properties": { "http": { "required": ["endpoint"] } },
				"required": ["http"]
			},
			"then": { "properties": { "datadog": { "not": { "required": ["site"] } } } }
		}
	]
}
`)

var _ = Schema.Add("TelemetryAuditLogStreamHTTPConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"endpoint": { "type": "string", "format": "x_http_url" }
	}
}
`)

var _ = Schema.Add("TelemetryAuditLogStreamDatadogConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"site": { "type": "string", "minLength": 1, "maxLength": 253 },
		"service": { "type": "string", "minLength": 1, "maxLength": 100 },
		"source": { "type": "string", "minLength": 1, "maxLength": 100 },
		"tags": {
			"type": "object",
			"maxProperties": 20,
			"additionalProperties": { "type": "string", "minLength": 1 }
		}
	}
}
`)
```

Four things about the clauses:

- The first four clauses are the matrix of [the stream object](../../specs/audit-log-streaming.md#the-stream-object), in both directions: `type` selects which encoding object is **allowed**, not merely which is required, so a `datadog` object on a syslog stream and a `syslog` object on a datadog stream are both rejected. That is what makes §1.4's materialisation rule observable, and it is why the two must land in the same commit.
- `datadog` and `http` are not in any `required` list. Every field of both has a default or is optional, so the objects themselves are optional, per the spec's "no, and only when …" rows.
- The fifth clause is `site` and `endpoint` being two ways to say the same thing. It is written once, as "not both", rather than as the spec's two symmetrical sentences. It has to be at the stream level because the two fields live in different child objects, and it has to be a schema rule rather than a `Validate` rule because schema validation runs on the raw document, before `SetDefaults` puts a `site` into every datadog stream.
- `tags` values are `minLength: 1`. Authgear does not rewrite a tag to fit Datadog's tag rules — the spec says Datadog normalises or rejects one that does not conform — but a pair whose value is empty produces a bare `key:` in `ddtags`, which is a malformed tag rather than a non-conforming one.
- `site` is **not** a closed enum of the nine sites Datadog documents today. It is a bare length-bounded string (`minLength: 1`, `maxLength: 253`, the DNS name limit), validated for shape only. Which sites exist is Datadog's deployment to govern, not this schema's: an enum would mean a new Datadog region, or one this project has never heard of, needs an Authgear release before a project can use it, and `LogsIntakeURL` (§1.2) builds the same URL shape for every site by string interpolation, so there is nothing for a closed set to protect against except a typo. A typo's failure mode is then delivery failing forever with a logged error — the same failure mode `http.endpoint` already has, and no worse a hazard than the escape hatch that already exists for exactly this "site not in the list" case.

### 1.6 `x_http_url` format — `pkg/lib/config/formats.go`

`http.endpoint` is an absolute `http` or `https` URL with a path, which no registered format covers. `http_origin` (`pkg/util/validation/formats.go:39`) is scheme + host with no path; `x_public_https_url` (line 66) is https-only and is used for `logo_uri`, a URL a browser loads; and `x_hook_uri` (line 57), which is otherwise the right shape, also accepts `authgeardeno:`, which has no meaning here.

Add to the existing `func init()` in `pkg/lib/config/formats.go`:

```go
jsonschemaformat.DefaultChecker["x_http_url"] = FormatHTTPURL{}
```

```go
// FormatHTTPURL checks that the input is an absolute http or https URL.
//
// http is accepted, as it is for a hook URL (x_hook_uri): the endpoint
// this validates is most often a Datadog Agent on loopback or a worker in
// the same cluster, and neither necessarily terminates TLS. What plaintext
// costs here is in the spec's caveats -- the DD-API-KEY header goes on the
// wire in the clear -- and it is the project's call, not this checker's.
//
// It is deliberately not an SSRF check either: whether the destination may
// be reached is decided at delivery time by the fetch address policy
// (part 02 §4.2), because a host that resolves publicly today may not
// tomorrow.
type FormatHTTPURL struct{}

func (f FormatHTTPURL) CheckFormat(ctx context.Context, value any) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	u, err := url.Parse(str)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("expect http or https scheme")
	}
	if u.Host == "" {
		return fmt.Errorf("expect non-empty host")
	}
	if u.User != nil {
		return fmt.Errorf("expect no userinfo")
	}
	return nil
}
```

Userinfo is rejected because a credential in the URL would be carried in every request and would reach the logs; the `DD-API-KEY` header is where the credential belongs.

Following the other checkers in that file, a non-string value is not an error — `"type": "string"` in the schema enforces the type.

### 1.7 Secret config — `pkg/lib/config/secret_telemetry.go`

A second named slice beside `TelemetryAuditLogStreamTLSMaterials`, in the same file, with the same stream-name-keyed shape:

```go
var _ = SecretConfigSchema.Add("TelemetryAuditLogStreamDatadogCredentials", `
{
	"type": "array",
	"items": { "$ref": "#/$defs/TelemetryAuditLogStreamDatadogCredentialsItem" }
}
`)

type TelemetryAuditLogStreamDatadogCredentials []TelemetryAuditLogStreamDatadogCredentialsItem

var _ SecretItemData = &TelemetryAuditLogStreamDatadogCredentials{}

// Resolve returns the credentials declared for streamName.
func (c *TelemetryAuditLogStreamDatadogCredentials) Resolve(streamName string) (*TelemetryAuditLogStreamDatadogCredentialsItem, bool) {
	if c == nil {
		return nil, false
	}
	for idx := range *c {
		item := (*c)[idx]
		if item.StreamName == streamName {
			return &item, true
		}
	}
	return nil, false
}

// SensitiveStrings returns every api_key, so that a key cannot appear in a
// log line. This differs from TelemetryAuditLogStreamTLSMaterials, which
// returns nil: a certificate is not a bearer credential, an API key is.
func (c *TelemetryAuditLogStreamDatadogCredentials) SensitiveStrings() []string {
	if c == nil {
		return nil
	}
	var out []string
	for _, item := range *c {
		if item.APIKey != "" {
			out = append(out, item.APIKey)
		}
	}
	return out
}

var _ = SecretConfigSchema.Add("TelemetryAuditLogStreamDatadogCredentialsItem", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"stream_name": { "type": "string", "pattern": "^[a-zA-Z0-9_-]{1,63}$" },
		"api_key": { "type": "string", "minLength": 1 }
	},
	"required": ["stream_name", "api_key"]
}
`)

type TelemetryAuditLogStreamDatadogCredentialsItem struct {
	StreamName string `json:"stream_name,omitempty"`
	APIKey     string `json:"api_key,omitempty"`
}
```

`api_key` is `minLength: 1` and not a hex pattern. A Datadog API key is 32 hex characters today, but the spec's `endpoint` escape hatch means the receiving end may be a project's own worker that accepts something else, and a shape check here would reject a working deployment for no gain.

### 1.8 Secret key registration — `pkg/lib/config/secret.go`

In the `SecretKey` const block, beside the tls key:

```go
	// nolint: gosec
	TelemetryAuditLogStreamDatadogCredentialsKey SecretKey = "telemetry.audit_logs.streams.datadog"
```

In `secretItemKeys`:

```go
	TelemetryAuditLogStreamDatadogCredentialsKey: {"TelemetryAuditLogStreamDatadogCredentials", func() SecretItemData { return &TelemetryAuditLogStreamDatadogCredentials{} }},
```

Not added to `SecretKey.IsUpdatable`: the portal has no UI for streams, and the tls secret it sits beside is not updatable either. The secret is written by editing `authgear.secrets.yaml`, or through the Admin API resource update path that writes the whole file.

### 1.9 Cross-validation — `pkg/lib/config/secret.go`

A datadog stream and its API key are validated as a pair, in `SecretConfig.Validate`, beside the SAML block it is modelled on (`secret.go:311-318`):

```go
	for _, stream := range appConfig.Telemetry.AuditLogs.Streams {
		if stream.Type == TelemetryAuditLogStreamTypeDatadog {
			c.validateTelemetryAuditLogStreamDatadogCredentials(vctx, stream.Name)
		}
	}
```

```go
func (c *SecretConfig) validateTelemetryAuditLogStreamDatadogCredentials(ctx *validation.Context, streamName string) {
	_, data, _ := c.LookupDataWithIndex(TelemetryAuditLogStreamDatadogCredentialsKey)
	credentials, _ := data.(*TelemetryAuditLogStreamDatadogCredentials)
	item, ok := credentials.Resolve(streamName)
	if !ok || item.APIKey == "" {
		ctx.EmitErrorMessage(fmt.Sprintf("api key of audit log stream '%s' is not configured", streamName))
	}
}
```

`credentials` may be nil here (the secret is absent entirely); `Resolve` is nil-safe, exactly as `SAMLSpSigningMaterials.Resolve` is at the call site above it. `appConfig.Telemetry.AuditLogs` is non-nil: `Parse` runs `SetFieldDefaults` before anything calls `Validate`, and neither field is nullable.

This placement is a departure from part 01 §1.7 of the syslog plan, which deliberately kept the tls secret out of `SecretConfig.Validate` because that function runs on every config **load** (`configsource/loader.go:49`) and an error there is a project that does not start. The difference is what the material is:

- tls material is optional. A stream with none still delivers, verifying against the system trust store, so an orphan is a tidiness problem.
- the API key is what makes a datadog stream able to deliver at all, and it is the credential, not a tuning knob. A `type: datadog` stream without one is the same kind of incomplete configuration as a SAML service provider without its certificates, which this function already refuses.

The consequence is stated in the spec and must be stated in release notes: a project left with a datadog stream and no key does not load. It is enforced on save for free — the portal's `ApplyUpdates0` calls `configsource.LoadConfig` on the new resource manager (`pkg/portal/appresource/manager.go:149`), which runs this validation — so no separate save-time check is written.

### 1.10 Orphan pruning — `pkg/lib/config/secret_update_instruction.go`

Pruning is opt-in, via a new `SecretConfigUpdateInstruction` field, not an automatic sweep on every `authgear.yaml` save. This replaces `pkg/portal/appresource/manager.go`'s `cleanupOrphanedSecrets`, which is deleted, for both keys — `telemetry.audit_logs.streams.tls` was already swept that way for the syslog feature, and the datadog key would otherwise be the only secret in the codebase pruned by an automatic scan rather than an explicit instruction.

The reason: every other secret's cleanup — `OAuthClientSecretsUpdateInstruction`'s `cleanup` action being the direct precedent — is something the caller who knows the new desired state asks for, by sending an instruction alongside the config write, not something the config layer infers by diffing against what used to be there. `authgear.secrets.yaml` is written exclusively through `AuthgearSecretYAMLDescriptor.UpdateResource` (`pkg/lib/config/configsource/resources.go`), which decodes its `data` as a `SecretConfigUpdateInstruction` and applies it — there is no code path where it is handed a raw replacement document, so an automatic sweep was already the odd one out structurally, not just for this key.

```go
type TelemetryAuditLogStreamSecretsUpdateInstructionCleanupData struct {
	KeepStreamNames []string `json:"keepStreamNames,omitempty"`
}

type TelemetryAuditLogStreamSecretsUpdateInstruction struct {
	Action SecretUpdateInstructionAction `json:"action,omitempty"`

	CleanupData *TelemetryAuditLogStreamSecretsUpdateInstructionCleanupData `json:"cleanupData,omitempty"`
}

func (i *TelemetryAuditLogStreamSecretsUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	switch i.Action {
	case SecretUpdateInstructionActionCleanup:
		return i.cleanup(currentConfig)
	default:
		return nil, fmt.Errorf("config: unexpected action for TelemetryAuditLogStreamSecretsUpdateInstruction: %s", i.Action)
	}
}
```

`cleanup` prunes **both** `telemetry.audit_logs.streams.tls` and `.datadog` in the one instruction, because both are keyed by the same `stream_name` and become orphaned by the same event — a stream leaving `telemetry.audit_logs.streams` — so the caller only has to compute the keep-set once:

```go
func pruneTelemetryAuditLogStreamSecretItems[T any](
	secretConfig *SecretConfig,
	key SecretKey,
	keep map[string]struct{},
	streamNameOf func(T) string,
) error
```

Same shape as the plan originally gave this helper, moved into `pkg/lib/config` since that is now the only package that needs it: `Lookup`, decode `item.RawData` (not `item.Data`, for the same `SetFieldDefaults`-materializes-a-panicking-`*JWK` reason as before), filter, drop the entry when nothing is left, otherwise re-marshal. Unlike the old version it does not special-case "nothing changed" — `OAuthClientSecretsUpdateInstruction.cleanup` always rewrites the item it touches, and this matches that.

What this changes operationally:

- **Nobody currently sends this instruction.** The portal has no UI for streams (§1.8), so unlike OAuth client removal — which the portal's own mutation drives, computing `keepClientIDs` from the form it just submitted — there is today no caller that both edits `telemetry.audit_logs.streams` and knows to send `telemetryAuditLogStreamSecrets: {action: cleanup, ...}` alongside it. An orphaned item left after removing a stream now stays until something explicitly asks for cleanup, which part 01 §5 decision 7 already called "tolerated" — it simply stays tolerated for longer, by default, than the old sweep left it.
- **Adding this instruction does not require adding portal UI.** It is reachable the same way every other instruction is reachable without portal UI today: a caller of the Admin API's resource-update endpoint that knows the desired stream set can send it directly.
- Pruning, when someone does send it, still has to happen with knowledge of the **new** `telemetry.audit_logs.streams`, i.e. after whatever write removed the stream — the same ordering constraint the old sweep had, just moved to whoever now issues the instruction instead of being enforced by `ApplyUpdates0` automatically.

### 1.11 Feature config

No change. The spec is explicit that the gate is on streaming as a whole and that a `type` is not gated on its own, so `telemetry.audit_logs.streaming.disabled` covers datadog streams with no new field and no new merge code.

## 2. Compatibility and deployment behaviour

- **No BREAKING-CHANGES entry.** Every new rule concerns configuration that cannot exist before this change: `type: datadog`, `transport: http`, and the two new objects. An existing `syslog` + `tcp` stream parses to the same struct values, defaults to the same values, and marshals to the same document as before §1.4 — the materialisation change is observable only for a combination that did not previously validate.
- **No storage, no migration, no API payload change.**
- **Rollout.** A binary older than this part rejects `type: datadog` as an enum violation and the `datadog` / `http` objects as unknown properties, because every object in this schema sets `additionalProperties: false`. A project must therefore not be given a datadog stream until every process serving it — including `authgear background`, which is what delivers — runs a binary with parts 01 and 02. Ordinary deploy-then-configure ordering.
- **Secret before stream.** Because of §1.9 the key must exist before, or with, the stream, and the stream must be removed before its key. This is the one operational ordering constraint the feature adds, and it is in the spec under [the datadog secret](../../specs/audit-log-streaming.md#the-datadog-secret).

## 3. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/config/telemetry.go` | `datadog`/`http` enum values, `DatadogSite` + `LogsIntakeURL`, the two new config structs and their schemas, `nullable:"true"` on the four object fields, `TelemetryAuditLogStreamConfig.SetDefaults`, `TelemetryAuditLogStreamTCPConfig.SetDefaults`, rewritten stream schema |
| `pkg/lib/config/formats.go` | `FormatHTTPURL` + `init` registration |
| `pkg/lib/config/secret_telemetry.go` | `TelemetryAuditLogStreamDatadogCredentials`, its item, `Resolve`, `SensitiveStrings`, schemas |
| `pkg/lib/config/secret.go` | secret key const, `secretItemKeys` entry, `validateTelemetryAuditLogStreamDatadogCredentials` + its loop in `Validate` |
| `pkg/lib/config/secret_update_instruction.go` | `TelemetryAuditLogStreamSecretsUpdateInstruction`, `pruneTelemetryAuditLogStreamSecretItems`, wired into `SecretConfigUpdateInstruction` |
| `pkg/portal/appresource/manager.go` | `cleanupOrphanedSecrets` and `pruneStreamKeyedSecret` **removed** — pruning moves to the instruction above, for both `tls` and `datadog` |
| `pkg/lib/config/telemetry_test.go` | defaulting and materialisation cases |
| `pkg/lib/config/secret_telemetry_test.go` | credentials parse / `Resolve` / `SensitiveStrings` cases |
| `pkg/lib/config/formats_test.go` | `FormatHTTPURL` cases |
| `pkg/lib/config/testdata/config_tests.yaml` | stream matrix and field validation cases |
| `pkg/lib/config/testdata/parse_secret_tests.yaml` | the new key in the expected-keys list, and parse cases |
| `pkg/lib/config/testdata/secret_config_validate_tests.yaml` | the datadog stream / api_key pair cases |
| `pkg/lib/config/testdata/secret_update_instruction.yaml` | `cleanup-telemetry-audit-log-stream-secrets-*` cases |
| `pkg/portal/appresource/manager_test.go` | the pre-existing tls pruning Convey block **removed** along with the code it tested |
| `docs/api/apis/*.json`, `portal/src/graphql/portal/globalTypes.generated.ts` | regenerated by `make export-schemas`, never hand-edited |

## 4. Test plan

`pkg/lib/config` uses Convey; every new Go test matches that style, per the `add-go-test` skill. The YAML fixtures are required by the `update-configs` skill — one per surface touched.

### 4.1 `pkg/lib/config/testdata/config_tests.yaml`

Cases, named `telemetry-audit-log-stream-datadog-*`, following the existing `telemetry-audit-log-stream-*` cases:

- a minimal datadog stream (`name`, `type: datadog`, `transport: http`, no objects) parses.
- a datadog stream with a full `datadog` object and an `http.endpoint`-free document parses.
- `type: datadog` + `transport: tcp` is rejected; `type: syslog` + `transport: http` is rejected.
- a `datadog` object on a `type: syslog` stream is rejected; a `syslog` object on a `type: datadog` stream is rejected.
- an `http` object on a `transport: tcp` stream is rejected; a `tcp` object on a `transport: http` stream is rejected.
- `datadog.site` and `http.endpoint` set together is rejected; each alone is accepted.
- `http.endpoint` accepts `https://opw.internal:8282/api/v2/logs` and the plaintext `http://localhost:8126/api/v2/logs`; rejects a value with no host, one carrying userinfo, and a non-http scheme.
- `datadog.site` accepts a site not in the nine Datadog documents today (`eu1.datadoghq.com`), proving it is not a closed enum, and accepts each of the nine; rejects the empty string and 254 characters.
- `datadog.service` and `datadog.source` reject the empty string and 101 characters.
- `datadog.tags` rejects 21 pairs and a pair with an empty value; accepts 20.
- two datadog streams with the same `name` are rejected by the existing uniqueness check, proving it is type-agnostic.

### 4.2 `pkg/lib/config/telemetry_test.go`

Extends the existing file, which already parses YAML through `parseTelemetryStreams`:

- a minimal datadog stream defaults `site` to `datadoghq.com`, `service` and `source` to `authgear`, and leaves `Tags` nil.
- explicit `site`, `service`, `source` and `tags` survive defaulting.
- **a datadog stream has `Syslog` and `TCP` nil after parsing**, and **a syslog stream has `Datadog` and `HTTP` nil** — the direct test of §1.4. Without it, a later removal of a `nullable` tag is caught only indirectly, by an effective-config round trip that nothing else exercises.
- a syslog stream still defaults `facility`, `app_name`, `structured_data_id` and `tcp.tls.enabled` — the existing cases, which must keep passing unchanged, since §1.4 rewrote how they are reached.
- `DatadogSite("datadoghq.eu").LogsIntakeURL()` is `https://http-intake.logs.datadoghq.eu/api/v2/logs`.

### 4.3 `pkg/lib/config/formats_test.go`

`FormatHTTPURL` accepts `https://http-intake.logs.datadoghq.com/api/v2/logs`, `https://opw.internal:8282/api/v2/logs` and `http://localhost:8126/api/v2/logs`; rejects `ftp://x/`, `authgeardeno:///x.ts` (which `x_hook_uri` would accept), `https://`, `/api/v2/logs` and `https://user:pw@x/`; returns nil for a non-string.

### 4.4 `pkg/lib/config/testdata/parse_secret_tests.yaml`

- `telemetry.audit_logs.streams.datadog` added to the expected-keys list in the unknown-secret case (line 11) — a mechanical update the suite fails without.
- a valid item parses; an item missing `api_key` is rejected; an item missing `stream_name` is rejected; an empty `api_key` is rejected; an unknown property is rejected.

### 4.5 `pkg/lib/config/testdata/secret_config_validate_tests.yaml`

The pair rule from §1.9:

- a `type: datadog` stream with a matching item validates.
- a `type: datadog` stream with **no** datadog secret at all fails with `api key of audit log stream 'X' is not configured`.
- a `type: datadog` stream whose secret has only another stream's item fails with the same message.
- two datadog streams, one keyed and one not, produce exactly one error, naming the unkeyed stream.
- a `type: syslog` stream with no datadog secret validates — the rule must not leak to the other type.
- a datadog secret item whose `stream_name` matches no stream validates: an orphan is tolerated, as the tls one is.

### 4.6 `pkg/lib/config/secret_telemetry_test.go`

- `Resolve` returns the matching item, `false` for an unknown name, and `false` on a nil receiver.
- `SensitiveStrings` returns every `api_key`, skips an empty one, and returns nil on a nil receiver.

### 4.7 `pkg/lib/config/testdata/secret_update_instruction.yaml`

`cleanup-telemetry-audit-log-stream-secrets-*` cases, alongside the existing `cleanup-oauth-client-secrets-*` ones the instruction is modelled on:

- a `cleanup` instruction whose `keepStreamNames` matches neither stream removes both the `tls` and `datadog` entries entirely, leaving every other secret byte-identical.
- one whose `keepStreamNames` matches one of two streams present in both keys prunes only the other stream's items, from both keys, in one instruction.
- one applied when only `telemetry.audit_logs.streams.datadog` exists (no `tls` entry at all) prunes it without erroring on the absent key.
- one applied when neither secret exists is a no-op, not an error.
- a `cleanup` instruction with no `cleanupData.keepStreamNames` is an error, naming the instruction.

`pkg/portal/appresource/manager_test.go` loses the "clean up orphaned audit log stream tls secrets" Convey block along with the code it tested; nothing replaces it there, because pruning is no longer something `ApplyUpdates0` does on its own.

### 4.8 Commands to run

```
go test ./pkg/lib/config/... ./pkg/portal/appresource/...
make export-schemas   # then confirm the regenerated files are committed
make lint
```

## 5. Fixed behavioural decisions

1. `type: datadog` pairs with `transport: http` only, and the schema enforces the matrix in both directions: an encoding or transport object that does not match the stream's `type`/`transport` is rejected, not ignored.
2. The four per-stream objects are `nullable:"true"` and materialised by `TelemetryAuditLogStreamConfig.SetDefaults` according to `type` and `transport`. An existing syslog stream's parsed and defaulted shape is unchanged.
3. `datadog.site` and `http.endpoint` are mutually exclusive, enforced in the schema because the check must see the raw document, before defaulting.
4. `http.endpoint` accepts both `http` and `https`, as a hook URL does; a plaintext endpoint is the project's call and its cost is in the spec's caveats. It is validated for shape (`x_http_url`) only, not against private or link-local ranges; the fetch address policy decides reachability at delivery time (part 02 §4.2).
5. The API key lives in `telemetry.audit_logs.streams.datadog`, keyed by stream name, and is **not** portal-updatable.
6. A `type: datadog` stream without an `api_key` is invalid **at load**, via `SecretConfig.Validate`, following the SAML service provider precedent rather than the tls secret's tolerate-and-prune treatment. The key is written before the stream and removed after it.
7. An orphaned datadog item is ignored at load, exactly as a tls item is, and pruned only when a caller sends `TelemetryAuditLogStreamSecretsUpdateInstruction` (action `cleanup`), which prunes both the `tls` and `datadog` keys together. This is not automatic on every `authgear.yaml` save -- that sweep is removed for `tls` too, in favour of the same opt-in pattern `OAuthClientSecretsUpdateInstruction`'s `cleanup` action already uses.
8. `TelemetryAuditLogStreamDatadogCredentials.SensitiveStrings` returns the keys, unlike the tls materials' nil.
9. No feature config change: streaming is gated as a whole, not per type.
10. `datadog.site` is a free-form, length-bounded string, not a closed enum of Datadog's documented sites. Which sites exist is Datadog's deployment to govern; a site this code has never heard of still works, and a typo fails at delivery time rather than at save time — the same trade `http.endpoint` already makes.

## 6. Atomic commit plan

Each commit builds and passes `go test ./pkg/lib/config/...` on its own.

### Commit 1 — `[Telemetry] Add the datadog type and http transport to the stream config`

- `pkg/lib/config/telemetry.go`, `pkg/lib/config/formats.go`
- `pkg/lib/config/telemetry_test.go`, `pkg/lib/config/formats_test.go`
- `pkg/lib/config/testdata/config_tests.yaml`
- regenerated schema artifacts from `make export-schemas`, **in this commit**

The enum values, the two objects, the materialisation rule and the schema matrix go together: the schema rejects what the materialisation rule stops `SetFieldDefaults` from producing, so splitting them leaves the tree in a state where a parsed syslog stream no longer round-trips.

### Commit 2 — `[Telemetry] Add the telemetry.audit_logs.streams.datadog secret`

- `pkg/lib/config/secret_telemetry.go`, `pkg/lib/config/secret.go`
- `pkg/lib/config/secret_telemetry_test.go`
- `pkg/lib/config/testdata/parse_secret_tests.yaml`, `secret_config_validate_tests.yaml`
- regenerated secret schema artifacts, in this commit

Carries both the type and the §1.9 pair rule. They are one behaviour — a key that is required is a key that is validated — and a bisect that lands between them would produce a binary that accepts a keyless datadog stream.

### Commit 3 — `[Telemetry] Prune telemetry audit log stream secrets via an explicit cleanup instruction`

- `pkg/lib/config/secret_update_instruction.go`, `pkg/lib/config/testdata/secret_update_instruction.yaml`
- `pkg/portal/appresource/manager.go`, `pkg/portal/appresource/manager_test.go` (the automatic sweep and its test removed, for both `tls` and `datadog`)

Not a datadog-only change: this replaces the syslog feature's automatic `cleanupOrphanedSecrets` sweep with an opt-in instruction, following `OAuthClientSecretsUpdateInstruction`'s `cleanup` action, for both keys at once. It lands as its own commit because it changes pre-existing `tls` behaviour, not just adds `datadog` pruning — a bisect on "orphaned secrets are no longer pruned automatically" should land here, not on a commit whose subject only mentions datadog.
