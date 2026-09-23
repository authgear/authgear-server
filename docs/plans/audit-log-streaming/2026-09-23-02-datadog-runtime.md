# Audit Log Streaming, type datadog — Part 02: Runtime

Implements delivery for `type: datadog` over `transport: http`, on top of [part 01](./2026-09-23-01-datadog-config.md) and the shipped syslog runtime ([2026-09-18-02-runtime](./2026-09-18-02-runtime.md)).

## Goal / scope

After this part a batch drained for a project with a datadog stream is encoded as Datadog log objects and POSTed, gzipped, to the logs HTTP intake.

Unchanged by this part, and deliberately so: the producer, the Redis queue and its scripts, `Runnable`, the drain lock, the per-app and per-stream concurrency bounds, and the whole syslog path. A stream's `type` is read for the first time inside `SenderImpl.sendToStream`; everything above it stays type-agnostic.

Part 03 adds e2e coverage.

## 1. Where the type is dispatched

`SenderImpl.Send` already fans out over `s.Streams` with a bounded pool (`transport.go:112-122`). `sendToStream` resolves the stream and then, today, dials TCP and writes syslog frames.

That body is split in two, and the switch goes between them:

```go
func (s *SenderImpl) sendToStream(
	ctx context.Context,
	logger slogutil.NamedLogger,
	streamConfig *config.TelemetryAuditLogStreamConfig,
	entries []QueuedEntry,
) {
	resolved, err := resolveStream(s.AppID, streamConfig, s.TLS, s.DatadogCredentials)
	if err != nil {
		logger.WithError(err).Error(ctx, "failed to resolve audit log stream",
			slog.String("app_id", s.AppID),
			slog.String("stream", streamConfig.Name))
		return
	}

	switch streamConfig.Type {
	case config.TelemetryAuditLogStreamTypeSyslog:
		s.sendSyslogTCP(ctx, logger, resolved, entries)
	case config.TelemetryAuditLogStreamTypeDatadog:
		s.sendDatadogHTTP(ctx, logger, resolved, entries)
	default:
		logger.Error(ctx, "unknown audit log stream type",
			slog.String("app_id", s.AppID),
			slog.String("stream", streamConfig.Name),
			slog.String("type", string(streamConfig.Type)))
	}
}
```

`sendSyslogTCP` is the current body of `sendToStream` from the `dial` call onwards, moved verbatim.

The switch is on `type`, not on `transport`, because `type` is the spec's primary discriminator and each case owns the one transport its type is paired with. The `default` logs rather than panics: a stream reaching here with an unknown type means a config schema and a sender that disagree, which must degrade one stream, not take down the tick — unlike `Frame`'s panic on an unknown framing, which is reachable only from a value this package itself produced.

## 2. Resolved runtime types — `resolve.go`

`ResolvedStream` grows the two axes as nullable sub-structs, so that "which encoding" and "which transport" are answered by a nil check rather than by re-reading the config:

```go
type ResolvedStream struct {
	AppID string
	Name  string

	// Exactly one of Syslog and Datadog is non-nil: the encoding the
	// stream's type selects.
	Syslog  *ResolvedSyslog
	Datadog *ResolvedDatadog

	// Exactly one of TCP and HTTP is non-nil: the delivery its transport
	// selects.
	TCP  *ResolvedTCP
	HTTP *ResolvedHTTP
}

type ResolvedTCP struct {
	Address    string
	TLSEnabled bool
	// TLSClientCertificate is nil unless the stream declares client_certificate.
	TLSClientCertificate *tls.Certificate
	// TLSRootCAs is nil unless the stream declares certificate_authority,
	// in which case the system trust store is not consulted.
	TLSRootCAs *x509.CertPool
}

type ResolvedHTTP struct {
	// Endpoint is the absolute URL every chunk is POSTed to, already
	// derived from the encoding object when http.endpoint is unset.
	Endpoint string
	// RootCAs carries the tls secret's certificate_authority, for a
	// self-hosted endpoint with a private CA. nil means the system trust
	// store. client_certificate is not used by this transport.
	RootCAs *x509.CertPool
}

type ResolvedDatadog struct {
	APIKey  string
	Service string
	Source  string
	// Tags are datadog.tags, sorted by key, with app_id and activity_type
	// already removed -- those two are appended per entry, from the event.
	Tags []DatadogTag
}

type DatadogTag struct {
	Key   string
	Value string
}
```

`ResolvedStream.Syslog` was a value (`Syslog ResolvedSyslog`) and the TCP fields were inline; both move into pointers. `sendSyslogTCP` then reads `resolved.Syslog` and `resolved.TCP`, and `dial` takes `*ResolvedTCP` instead of `*ResolvedStream`. Mechanical, and it is what makes an encoding/transport mismatch impossible to express.

### 2.1 `resolveStream`

```go
func resolveStream(
	appID string,
	streamConfig *config.TelemetryAuditLogStreamConfig,
	materials *config.TelemetryAuditLogStreamTLSMaterials,
	credentials *config.TelemetryAuditLogStreamDatadogCredentials,
) (*ResolvedStream, error)
```

In order:

1. `resolved := &ResolvedStream{AppID: appID, Name: streamConfig.Name}`.
2. `item, _ := materials.Resolve(streamConfig.Name)` — looked up once, used by whichever transport branch runs. `Resolve` is nil-safe.
3. Encoding, on `streamConfig.Type`:
   - `syslog` → `resolved.Syslog = &ResolvedSyslog{…}` from `streamConfig.Syslog`, exactly as today, including `Facility: streamConfig.Syslog.Facility.Code()`.
   - `datadog` → `resolved.Datadog, err = resolveDatadog(streamConfig, credentials)`.
   - default → `fmt.Errorf("auditlogstreaming: unknown stream type %q for stream %q", …)`.
4. Transport, on `streamConfig.Transport`:
   - `tcp` → `resolved.TCP = &ResolvedTCP{Address: …, TLSEnabled: …}` plus the existing certificate work, unchanged: CA pool when `item.CertificateAuthority != nil`, client certificate when `item.ClientCertificate != nil`, both skipped entirely when `TLSEnabled` is false.
   - `http` → `resolved.HTTP, err = resolveHTTP(streamConfig, item)`.
   - default → an error, as above.

Part 01 §1.4 guarantees `streamConfig.Syslog`/`TCP` are non-nil for a syslog/tcp stream and `Datadog`/`HTTP` are non-nil for a datadog/http stream, so each branch reads its own objects without nil guards. It must not read the other pair's, which is now nil rather than an empty struct.

```go
func resolveDatadog(
	streamConfig *config.TelemetryAuditLogStreamConfig,
	credentials *config.TelemetryAuditLogStreamDatadogCredentials,
) (*ResolvedDatadog, error) {
	item, ok := credentials.Resolve(streamConfig.Name)
	if !ok || item.APIKey == "" {
		return nil, fmt.Errorf("auditlogstreaming: missing api key for datadog stream %q", streamConfig.Name)
	}
	…
}
```

The missing-key branch is defensive: part 01 §1.9 makes a keyless datadog stream fail to load, so a project reaching here has one. It returns an error rather than sending without the header, because an unauthenticated POST would be a `403` per chunk and a wasted round trip per tick.

`Tags` are built here, once per tick, rather than per entry:

```go
	keys := make([]string, 0, len(streamConfig.Datadog.Tags))
	for k := range streamConfig.Datadog.Tags {
		// app_id and activity_type come from the event, and the spec drops
		// a configured pair that would collide with them.
		if k == "app_id" || k == "activity_type" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
```

Sorted, because Go's map iteration order is randomised and an unstable `ddtags` would make every golden test flaky and every log line differ from the last for no reason. The spec does not fix an order for the configured pairs; this plan does.

```go
func resolveHTTP(
	streamConfig *config.TelemetryAuditLogStreamConfig,
	item *config.TelemetryAuditLogStreamTLSMaterialsItem,
) (*ResolvedHTTP, error)
```

- `Endpoint`: `streamConfig.HTTP.Endpoint` when non-empty; otherwise derived from the encoding object — for a datadog stream, `streamConfig.Datadog.Site.LogsIntakeURL()` (part 01 §1.2). Part 01 §1.5's schema guarantees the two are never both set, so there is no precedence rule to get wrong; the `if` is which one is present, not which one wins. A future `type` paired with `http` adds a case here.
- `RootCAs`: when `item != nil && item.CertificateAuthority != nil`, `x509.NewCertPool()` + `AppendCertsFromPEM`, erroring when nothing was appended — the same code the tcp branch runs. `item.ClientCertificate` is ignored: the spec says this transport does not present one.
- Unlike the tcp branch, there is no `TLSEnabled` gate: the endpoint's scheme decides. `RootCAs` is resolved whenever the tls secret declares a `certificate_authority`, and it simply goes unused for an `http` endpoint, where `net/http` never builds a TLS connection. Nothing needs to detect that case — resolving a pool for a plaintext stream costs one parse per tick and keeps this branch free of a scheme check that would then have to agree with the one in `net/http`.

## 3. The log object — `datadog.go` (new)

```go
func EncodeDatadogLog(d *ResolvedDatadog, hostname string, entry QueuedEntry) ([]byte, error)
func datadogStatus(t event.Type) string
func buildDDTags(tags []DatadogTag, appID string, activityType event.Type) string
```

`EncodeDatadogLog` marshals one struct, so the field order is fixed by the struct and every omission is an `omitempty`:

```go
type datadogLog struct {
	DDSource  string `json:"ddsource"`
	Service   string `json:"service"`
	Hostname  string `json:"hostname,omitempty"`
	DDTags    string `json:"ddtags,omitempty"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`

	Evt     *datadogEvt     `json:"evt,omitempty"`
	Usr     *datadogUsr     `json:"usr,omitempty"`
	Network *datadogNetwork `json:"network,omitempty"`
	HTTP    *datadogHTTP    `json:"http,omitempty"`

	Authgear datadogAuthgear `json:"authgear"`
}

type datadogEvt struct {
	Name string `json:"name,omitempty"`
}

type datadogUsr struct {
	ID string `json:"id,omitempty"`
}

type datadogNetwork struct {
	Client *datadogNetworkClient `json:"client,omitempty"`
}

type datadogNetworkClient struct {
	IP    string               `json:"ip,omitempty"`
	GeoIP *datadogNetworkGeoIP `json:"geoip,omitempty"`
}

type datadogNetworkGeoIP struct {
	Country datadogNetworkCountry `json:"country"`
}

type datadogNetworkCountry struct {
	ISOCode string `json:"iso_code,omitempty"`
}

type datadogHTTP struct {
	UserAgent string `json:"useragent,omitempty"`
	URL       string `json:"url,omitempty"`
	Referer   string `json:"referer,omitempty"`
}

// datadogAuthgear carries the event object verbatim. Event is the bytes
// the producer enqueued (QueuedEntry.Raw), not a re-marshalling of
// QueuedEntry.Event: that struct has no Payload (see QueuedEntry's doc
// comment), so re-marshalling would silently drop the whole payload.
type datadogAuthgear struct {
	Event json.RawMessage `json:"event"`
}
```

Field sources, from [the log object](../../specs/audit-log-streaming.md#the-log-object):

| Field | Value |
|---|---|
| `ddsource` | `d.Source` |
| `service` | `d.Service` |
| `hostname` | `hostname`, omitted when empty (§3.1) |
| `ddtags` | `buildDDTags(d.Tags, e.Context.AppID, e.Type)` |
| `message` | `string(e.Type)` |
| `status` | `datadogStatus(e.Type)` |
| `timestamp` | `e.Context.Timestamp * 1000` |
| `evt.name` | `string(e.Type)` |
| `usr.id` | `*e.Context.UserID`, the whole `usr` object omitted when the pointer is nil or the value empty |
| `network.client.ip` | `e.Context.IPAddress` |
| `network.client.geoip.country.iso_code` | `*e.Context.GeoLocationCode`, omitted when nil or empty |
| `http.useragent` | `e.Context.UserAgent` |
| `http.url` | `e.Context.AuditContext["http_url"]`, when present and a string |
| `http.referer` | `e.Context.AuditContext["http_referer"]`, when present and a string |
| `authgear.event` | `entry.Raw` |

`AuditContext` is `map[string]any` (`pkg/api/event/context.go:54`), so each lookup is a comma-ok type assertion to `string`; anything else is treated as absent, per the spec. `http_url` is written by `event.NewAuditContext`; `http_referer` only by the portal and Site Admin API paths, which is why the spec marks it as sometimes absent.

A nested object is emitted only when at least one of its own fields survived: `network` is nil when both the IP and the country code are empty, and `http` is nil when all three are empty. Building the object with three empty strings and letting `omitempty` clear the leaves would still emit `"network":{"client":{}}`, which becomes a facet-less empty attribute in Datadog.

`datadogStatus` shares the syslog severity list rather than repeating it, which is what makes the spec's "the two types classify an entry the same way" true by construction:

```go
// datadogStatus maps an activity type to a Datadog status. It reads the
// same severityByType table PRI does (encoder.go), so adding an activity
// type to that table changes both types' classification at once.
func datadogStatus(t event.Type) string {
	if _, ok := severityByType[t]; ok {
		return "warning"
	}
	return "info"
}
```

`buildDDTags` writes `key:value` pairs joined by `,`: every element of `tags` in order, then `app_id:<appID>` when `appID` is non-empty, then `activity_type:<type>`. Nothing is escaped or normalised — the spec is explicit that a tag which does not conform to Datadog's rules is Datadog's to normalise or reject.

### 3.1 Hostname — `hostname.go`

`hostname` is omitted when the project's `public_origin` has no usable host, whereas syslog's HOSTNAME must be the NILVALUE `-`. The "-" substitution therefore moves out of the resolver and into the syslog encoder:

```go
// ResolveHostname returns the host of a project's http.public_origin,
// without the port, or "" when there is none usable. It identifies the
// project the entries belong to, not the Authgear process that delivers
// them.
func ResolveHostname(publicOrigin string) string   // "" instead of "-"
```

`sanitizeHostname` returns `""` instead of `"-"`, and `EncodeRFC5424` writes `-` when the hostname it is given is empty. `NewSenderImpl` keeps one `Hostname` field; the two encoders differ in how they spell "absent", which is where that difference belongs.

## 4. Delivery — `datadog_transport.go` (new)

### 4.1 Constants

```go
const (
	// Datadog's documented intake limits: at most 1000 logs in the array,
	// and 5MB of uncompressed body. A batch is split into as many requests
	// as these require.
	datadogMaxEntriesPerRequest = 1000
	datadogMaxRequestBytes      = 5 * 1000 * 1000

	// datadogMaxLogBytes is the largest encoded log that can be sent at
	// all: one element alone in the array, minus the two brackets and the
	// separator fits() charges unconditionally. Deriving it from the body
	// limit, rather than writing a second number, is what guarantees that
	// any log the sender keeps fits an empty chunk -- see §4.3 step 4b.
	datadogMaxLogBytes = datadogMaxRequestBytes - 3
)

// datadogRequestTimeout bounds one chunk's request; datadogBatchTimeout
// bounds every chunk of one stream's batch together, so that a destination
// that is slow rather than down cannot hold the drain lock (runnable.go,
// WithMutexExpiry) for chunks * datadogRequestTimeout. Both are var so
// tests can shrink them.
var (
	datadogRequestTimeout = 30 * time.Second
	datadogBatchTimeout   = 2 * time.Minute
)
```

`datadogBatchTimeout` is the http analogue of the tcp path's single `SetWriteDeadline` for the whole batch (`transport.go:151`): one deadline for the transfer, not one per write.

### 4.2 The client

The endpoint comes from a project's configuration, so it is fetched with the client every other project-chosen URL uses, `httputil.NewSSRFSafeExternalClient`, carrying the deployment's `http.insecure_fetch_address_allowed` and `http.insecure_fetch_address_allowed_hosts`. A project cannot then point a stream at the deployment's internal network; an operator who genuinely runs an Agent or an Observability Pipelines Worker there allows the host in `authgear.features.yaml`, which only they can write.

The same client serves both schemes. An `http` endpoint is accepted (part 01 §1.6, as for a hook URL), and for it `net/http` never builds a TLS connection, so the `RootCAs` below and the TLS floor simply do not come into play. The address policy still does: scheme and address are independent, and a plaintext endpoint on a non-public address needs the operator's flag exactly as an https one does.

The tls secret's `certificate_authority` has to reach that client, and `SSRFSafeExternalClientOptions` has no way to carry it today. `pkg/util/httputil/ext_client.go` gains one field:

```go
type SSRFSafeExternalClientOptions struct {
	AllowNonPublicAddresses bool
	AllowedHosts            []string
	Sink                    string
	// RootCAs verifies the server certificate against this pool only,
	// instead of the system trust store. For a destination inside the
	// deployment, issued by a private certificate authority. nil keeps the
	// system trust store.
	RootCAs *x509.CertPool
}
```

and `NewSSRFSafeExternalClient` sets, on the `http.Transport` it already builds:

```go
	TLSClientConfig: &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    opts.RootCAs,
	},
```

Set unconditionally, not only when `RootCAs != nil`. `tls.VersionTLS12` is already Go's client default, so no existing caller's handshake changes, and stating it makes the spec's "at least 1.2" a property of the code rather than of the toolchain version. `ForceAttemptHTTP2` stays `true`, which is what keeps HTTP/2 enabled now that `TLSClientConfig` is non-nil.

Because the pool is per stream, the client is built per stream, not per sender:

```go
// DatadogClientFactory builds the client for one datadog stream. It is an
// interface so that a test can hand the sender a client pointed at an
// httptest server without minting a certificate authority.
type DatadogClientFactory interface {
	MakeClient(rootCAs *x509.CertPool) *http.Client
}

type SSRFSafeDatadogClientFactory struct {
	AllowNonPublicAddresses bool
	AllowedHosts            []string
}

func (f *SSRFSafeDatadogClientFactory) MakeClient(rootCAs *x509.CertPool) *http.Client {
	return httputil.NewSSRFSafeExternalClient(datadogRequestTimeout, httputil.SSRFSafeExternalClientOptions{
		AllowNonPublicAddresses: f.AllowNonPublicAddresses,
		AllowedHosts:            f.AllowedHosts,
		Sink:                    "telemetry.audit_logs.streams",
		RootCAs:                 rootCAs,
	})
}
```

`Sink` names the configuration that chose the URL, so a refusal logs which setting to fix.

A client per stream per tick means a fresh `http.Transport` and no connection reuse between ticks. At one delivery a minute that is the same trade the tcp transport already makes by dialling per tick, and it keeps the per-stream CA from leaking into another stream's connection pool.

### 4.3 `sendDatadogHTTP`

```go
func (s *SenderImpl) sendDatadogHTTP(
	ctx context.Context,
	logger slogutil.NamedLogger,
	resolved *ResolvedStream,
	entries []QueuedEntry,
)
```

Once per datadog stream per tick:

1. `ctx, cancel := context.WithTimeout(ctx, datadogBatchTimeout)`, `defer cancel()`.
2. `client := s.DatadogClientFactory.MakeClient(resolved.HTTP.RootCAs)`.
3. `w := newDatadogChunkWriter()`.
4. For each entry, in order:
   a. `log, err := EncodeDatadogLog(resolved.Datadog, s.Hostname, entry)`. On error, log one warning with the event ID and skip the entry — one unencodable entry must not cost the batch.
   b. `if len(log) > datadogMaxLogBytes`, log one warning naming the event ID and its size, and skip it: the spec's "a single entry that alone exceeds the body limit is dropped", and it cannot be delivered by any splitting. This threshold must stay derived from the same arithmetic `fits` uses; a log that passes here and then fails `fits` on a fresh chunk would be written into an over-limit request anyway, since step d adds unconditionally.
   c. ```go
      if !w.fits(log) {
          if w.count > 0 {
              if !s.postChunk(ctx, …, w.finish()) {
                  return
              }
          }
          w = newDatadogChunkWriter()
      }
      ```
      The `w.count > 0` guard is what stops an empty `[]` from being POSTed. It is unreachable while step b holds — `fits` is always true for a kept log on a fresh writer — and it is written anyway so that a later change to either threshold degrades into a wasted request rather than a stream of empty arrays.
   d. `w.add(log)`.
5. `if w.count > 0 { s.postChunk(ctx, …, w.finish()) }`.
6. Log one Debug line with the stream, the entry count and the chunk count.

`postChunk` returns false when the rest of the batch must be abandoned; see §4.5.

Entries are encoded one at a time into the chunk under construction rather than all up front. A 10000-entry batch at a few kilobytes each is tens of megabytes; holding one chunk plus one log instead is the same bound the tcp path takes with `maxEntriesPerWrite`.

### 4.4 `datadogChunkWriter`

```go
// datadogChunkWriter accumulates encoded logs into one gzipped JSON array,
// tracking the uncompressed size so that Datadog's limit -- which is on
// the uncompressed body -- can be respected while only the compressed
// bytes are held.
type datadogChunkWriter struct {
	buf          bytes.Buffer
	gz           *gzip.Writer
	count        int
	uncompressed int
}

func newDatadogChunkWriter() *datadogChunkWriter {
	w := &datadogChunkWriter{}
	w.gz = gzip.NewWriter(&w.buf)
	_, _ = w.gz.Write([]byte("["))
	w.uncompressed = 1
	return w
}

// fits reports whether log can be added without exceeding either limit.
// The two extra bytes are the comma that would separate log from the
// previous element and the closing "]". The comma is charged even for the
// first element, where there is none, so the bound is one byte
// conservative -- cheaper than branching in both fits and add, for a byte
// out of five million.
func (w *datadogChunkWriter) fits(log []byte) bool {
	return w.count < datadogMaxEntriesPerRequest &&
		w.uncompressed+1+len(log)+1 <= datadogMaxRequestBytes
}

func (w *datadogChunkWriter) add(log []byte) error {
	if w.count > 0 {
		if _, err := w.gz.Write([]byte(",")); err != nil {
			return err
		}
		w.uncompressed++
	}
	if _, err := w.gz.Write(log); err != nil {
		return err
	}
	w.uncompressed += len(log)
	w.count++
	return nil
}

func (w *datadogChunkWriter) finish() ([]byte, error) {
	if _, err := w.gz.Write([]byte("]")); err != nil {
		return nil, err
	}
	// Close flushes gzip's own buffer; buf is short of the trailer until
	// it returns, so buf.Bytes() is only valid afterwards.
	if err := w.gz.Close(); err != nil {
		return nil, err
	}
	return w.buf.Bytes(), nil
}
```

Three properties this shape has, and which the chunking depends on:

- **The size that is tracked is the uncompressed one**, accumulated arithmetically from `len(log)` and the separators, because Datadog's 5MB limit is on the uncompressed body. `buf.Len()` is never consulted: it is the compressed length, and it lags anyway until `Close`. Tracking uncompressed is also conservative against a compressed cap, since gzip does not expand JSON of this shape.
- **Each log is encoded exactly once.** `fits` is pure arithmetic on a length, so nothing is encoded, measured and thrown away, and a log that does not fit is carried into the next chunk as the same `[]byte`.
- **`fits` is checked before `add`, never after.** A `gzip.Writer` is append-only, so a log written into an over-full chunk cannot be taken back out; the pre-check is the only place the decision can be made.

The body is an array even for a single entry. The intake also accepts a bare object; one code path is worth more than the two saved bytes.

### 4.5 `postChunk`

```go
// postChunk sends one chunk and reports whether the batch should continue.
func (s *SenderImpl) postChunk(
	ctx context.Context,
	logger slogutil.NamedLogger,
	client *http.Client,
	resolved *ResolvedStream,
	body []byte,
	count int,
) bool
```

1. `req, err := http.NewRequestWithContext(ctx, http.MethodPost, resolved.HTTP.Endpoint, bytes.NewReader(body))`.
2. Headers: `DD-API-KEY: <resolved.Datadog.APIKey>`, `Content-Type: application/json`, `Content-Encoding: gzip`.
3. `resp, err := client.Do(req)`. On error — a refusal by the fetch address policy, a TLS failure, a timeout — log one error with the stream and the error, and return `true`: this chunk is dropped, the remaining chunks are still sent, per the spec.
4. `defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()`, so the connection can be reused for the next chunk. The client already caps the body at `httputil.MaxResponseBytes`.
5. `resp.StatusCode == http.StatusAccepted` → Debug log with the chunk's entry count, return `true`.
6. `resp.StatusCode == 401 || 403` → read up to 1KB of the body, log one error saying the API key is missing, wrong or revoked, and return **false**. Every remaining chunk carries the same key and would fail identically; sending them would turn one configuration error into one request per thousand entries.
7. Any other status → read up to 1KB of the body, log one error with the status and that snippet, return `true`.

Nothing is retried, and the queue entries are already gone (`drainScript` clears on take), so a dropped chunk is a gap. That is the spec's at-most-once contract; [retry](../../specs/audit-log-streaming.md#retry) is future work.

`202` means accepted, not stored as intended — the intake answers `202` to a malformed body too — so a success log is not evidence the encoder is correct. That is what part 03 exists for.

## 5. Sender construction and wiring

### 5.1 `SenderImpl`

```go
type SenderImpl struct {
	AppID    string
	Hostname string
	Streams  []*config.TelemetryAuditLogStreamConfig
	TLS      *config.TelemetryAuditLogStreamTLSMaterials

	DatadogCredentials   *config.TelemetryAuditLogStreamDatadogCredentials
	DatadogClientFactory DatadogClientFactory
}
```

```go
func NewSenderImpl(
	appID string,
	httpConfig *config.HTTPConfig,
	httpFeatureConfig *config.HTTPFeatureConfig,
	telemetryConfig *config.TelemetryConfig,
	tlsMaterials *config.TelemetryAuditLogStreamTLSMaterials,
	datadogCredentials *config.TelemetryAuditLogStreamDatadogCredentials,
) *SenderImpl {
	return &SenderImpl{
		AppID:              appID,
		Hostname:           ResolveHostname(httpConfig.PublicOrigin),
		Streams:            telemetryConfig.AuditLogs.Streams,
		TLS:                tlsMaterials,
		DatadogCredentials: datadogCredentials,
		DatadogClientFactory: &SSRFSafeDatadogClientFactory{
			AllowNonPublicAddresses: httpFeatureConfig.IsInsecureFetchAddressAllowed(),
			AllowedHosts:            httpFeatureConfig.GetInsecureFetchAddressAllowedHosts(),
		},
	}
}
```

Both accessors are nil-safe on `*HTTPFeatureConfig`, as `pkg/lib/hook/http.go` relies on.

### 5.2 Providers

`pkg/lib/deps/deps_config.go`, beside `ProvideTelemetryAuditLogStreamTLSMaterials` (line 287) and added to the same `secretDeps` set (line 150):

```go
func ProvideTelemetryAuditLogStreamDatadogCredentials(c *config.SecretConfig) *config.TelemetryAuditLogStreamDatadogCredentials {
	s, _ := c.LookupData(config.TelemetryAuditLogStreamDatadogCredentialsKey).(*config.TelemetryAuditLogStreamDatadogCredentials)
	return s
}
```

`*config.HTTPFeatureConfig` needs no new provider: `"HTTP"` is already in `wire.FieldsOf(new(*config.FeatureConfig), …)` (`deps_config.go:76`), and `newSenderImpl`'s graph reaches `*config.FeatureConfig` through `wire.FieldsOf(new(*config.Config), …)`.

`cmd/authgear/background/wire.go`'s `newSenderImpl` is unchanged — the two new parameters resolve through `DependencySet` — but `cmd/authgear/background/wire_gen.go` must be regenerated in the same commit.

Nothing on the producer side changes: `Producer.Enqueue` does not look at a stream's type.

## 6. Compatibility and deployment behaviour

- **No new persisted state.** No SQL, no migration, no new Redis key, and the queue payload is unchanged — a datadog stream reads the same `QueuedEntry` a syslog stream does.
- **Mixed-version rollout.** A queue entry written by any version is deliverable by any version; only the background worker needs parts 01 and 02 to deliver a datadog stream, and part 01 already keeps the config from existing before then.
- **No API change**, no BREAKING-CHANGES entry: the `httputil` change adds an option and pins a TLS floor that was already the default, and every other change is reachable only through configuration that could not previously exist.
- **Outbound traffic shape.** A datadog project's worker makes at least one request per stream per tick, against Datadog's public intake by default. [Rate limits](https://docs.datadoghq.com/api/latest/rate-limits/) states the logs intake is not rate limited, so chunks are not paced.
- **Egress.** A deployment that restricts outbound traffic must allow `http-intake.logs.<site>`; until then every chunk fails at `client.Do` and is logged.

## 7. File-level change plan

| File | Change |
|---|---|
| `pkg/util/httputil/ext_client.go` | `RootCAs` option, explicit `TLSClientConfig` with `MinVersion` |
| `pkg/lib/telemetry/auditlogstreaming/datadog.go` | **new** — the log object, `EncodeDatadogLog`, `datadogStatus`, `buildDDTags` |
| `pkg/lib/telemetry/auditlogstreaming/datadog_transport.go` | **new** — constants, `DatadogClientFactory`, `datadogChunkWriter`, `sendDatadogHTTP`, `postChunk` |
| `pkg/lib/telemetry/auditlogstreaming/resolve.go` | `ResolvedTCP`/`ResolvedHTTP`/`ResolvedDatadog`, `resolveStream` dispatch, `resolveDatadog`, `resolveHTTP` |
| `pkg/lib/telemetry/auditlogstreaming/transport.go` | type switch in `sendToStream`, `sendSyslogTCP` extracted, `dial` takes `*ResolvedTCP`, new `SenderImpl` fields |
| `pkg/lib/telemetry/auditlogstreaming/encoder.go` | `EncodeRFC5424` writes `-` for an empty hostname |
| `pkg/lib/telemetry/auditlogstreaming/hostname.go` | `ResolveHostname` returns `""` rather than `-` |
| `pkg/lib/telemetry/auditlogstreaming/deps.go` | `NewSenderImpl` signature and body |
| `pkg/lib/deps/deps_config.go` | `ProvideTelemetryAuditLogStreamDatadogCredentials`, added to `secretDeps` |
| `cmd/authgear/background/wire_gen.go` | regenerated by `make generate` |
| `pkg/util/httputil/ext_client_test.go` | `RootCAs` case |
| `pkg/lib/telemetry/auditlogstreaming/datadog_test.go` | **new** |
| `pkg/lib/telemetry/auditlogstreaming/datadog_transport_test.go` | **new** |
| `pkg/lib/telemetry/auditlogstreaming/resolve_test.go`, `transport_test.go`, `encoder_test.go`, `hostname_test.go` | extended / adjusted |

## 8. Test plan

Convey throughout, matching the package's existing tests.

### 8.1 `datadog_test.go`

Pure, no I/O, asserting on the exact JSON.

- a `user.authenticated` entry with every context field populated produces exactly the object in [the spec's example](../../specs/audit-log-streaming.md#example), with `authgear.event` byte-identical to `QueuedEntry.Raw` — the case that proves the payload is carried verbatim rather than re-marshalled from the payload-less `QueuedEntry.Event`.
- `timestamp` is `Context.Timestamp * 1000` and ends in `000`; it reflects the event, not the time of encoding.
- a nil `Context.UserID` omits the whole `usr` object; an empty `IPAddress` with a nil `GeoLocationCode` omits the whole `network` object; empty user agent, url and referer omit the whole `http` object.
- `http.url` and `http.referer` are taken from `AuditContext`, and an entry whose `audit_context` holds a non-string under those keys omits them rather than erroring.
- `hostname` is omitted when the resolved hostname is empty, and present otherwise.
- `datadogStatus` is `warning` for `email.error`, `sms.error`, `whatsapp.error`, and `info` for `user.authenticated` and an invented `some.new.type` — the same three the severity test pins, deliberately duplicated so that a change to one classification fails both.
- `buildDDTags`: configured pairs come first in sorted order, then `app_id`, then `activity_type`; a configured `app_id` or `activity_type` pair is dropped (resolved out in `resolveDatadog`, asserted there); no tags at all still yields `app_id:…,activity_type:…`; an empty `Context.AppID` omits the `app_id` tag.
- `ddsource`, `service`, `message`, `status`, `timestamp` and `authgear.event` are present even when every optional field is empty.

### 8.2 `datadog_transport_test.go` — against an `httptest.Server`

The factory is stubbed with `srv.Client()`, so these run without a certificate authority; §8.5 covers the real client.

- one entry produces one `POST`, whose body gunzips to a one-element array, with `DD-API-KEY`, `Content-Type: application/json` and `Content-Encoding: gzip` set.
- `datadogMaxEntriesPerRequest` shrunk to 3, 7 entries produce 3 requests of 3, 3 and 1, in occurrence order across requests.
- `datadogMaxRequestBytes` shrunk, entries split by size rather than by count, and no chunk's uncompressed body exceeds the limit.
- an entry alone larger than the limit is skipped with a warning, and the entries around it are still delivered — the case that proves one oversized entry does not poison the batch.
- the boundary, with `datadogMaxRequestBytes` shrunk: a log of exactly `datadogMaxLogBytes` is delivered, alone in its own chunk, and one byte larger is dropped. This is what pins step 4b's threshold to `fits`'s arithmetic; the two drifting apart produces either an over-limit request or an entry dropped for no reason.
- no request ever carries an empty array, including when the batch's only entries were all dropped for size — then no request is made at all.
- a `500` on the first of three chunks still sends the remaining two.
- a `403` on the first of three chunks sends **no** further chunk.
- a transport error (server closed) drops that chunk and continues.
- a server that never responds is abandoned when `datadogBatchTimeout` elapses, rather than hanging the tick.
- the endpoint is the `http.endpoint` verbatim when set, and `https://http-intake.logs.<site>/api/v2/logs` when it is not (asserted through `resolveHTTP`, §8.3).

### 8.3 `resolve_test.go`

- a datadog stream resolves `Datadog` non-nil and `Syslog` nil; a syslog stream the reverse; likewise `HTTP`/`TCP`.
- a datadog stream with no matching credentials item, and one whose item has an empty `api_key`, each return an error.
- `resolveHTTP` derives the intake URL from `site` for the default site and an arbitrary one (part 01 §1.5 makes `site` a free string, not an enum, so this proves the interpolation rather than a fixed list), and returns `http.endpoint` verbatim when it is set.
- `certificate_authority` in the tls secret becomes `ResolvedHTTP.RootCAs`; a malformed PEM is an error; `client_certificate` alone leaves `RootCAs` nil and is not otherwise used.
- configured tags are sorted by key, and a configured `app_id` or `activity_type` pair is dropped.
- the existing syslog/tcp resolution cases keep passing against the restructured `ResolvedStream`.

### 8.4 `transport_test.go`

- a project with one syslog stream and one datadog stream delivers to both, and a failure of either does not affect the other — the UC3 property across two types.
- a stream whose `type` is unknown is skipped with an error log and does not stop the other streams.

### 8.5 `pkg/util/httputil/ext_client_test.go`

- a client built with `RootCAs` set to a pool containing an `httptest.NewTLSServer`'s certificate completes the handshake; one built without it fails.
- `AllowNonPublicAddresses: false` still refuses a loopback address even when `RootCAs` is set — the regression that would turn the new option into an SSRF bypass.

### 8.6 Commands to run

```
go test ./pkg/lib/telemetry/... ./pkg/util/httputil/... ./pkg/lib/deps/... ./cmd/authgear/...
make generate   # wire_gen.go for the background graph
make lint
make test
```

## 9. Fixed behavioural decisions

1. Delivery dispatches on `type` inside `SenderImpl.sendToStream`. Nothing above it — the producer, the queue, `Runnable`, the concurrency bounds — is type-aware.
2. `ResolvedStream` carries one encoding sub-struct and one transport sub-struct, both nullable; the pairing is established once in `resolveStream`.
3. The endpoint is `http.endpoint` when set, otherwise derived from `datadog.site`. Part 01's schema makes both-set unreachable, so there is no precedence rule.
4. The request goes through `httputil.NewSSRFSafeExternalClient`, carrying the deployment's fetch address policy. A non-public endpoint requires `http.insecure_fetch_address_allowed` or an entry in `http.insecure_fetch_address_allowed_hosts`, whatever its scheme.
5. Both `http` and `https` endpoints go through the same client and the same code path; scheme and address are independent axes, and nothing in this package branches on the scheme.
6. The tls secret's `certificate_authority` applies to an `https` endpoint of the http transport, and is inert for an `http` one; `client_certificate` is not used by this transport at all.
7. One client per stream per tick, with that stream's CA pool. Nothing is pooled across ticks.
8. The body is always a gzipped JSON array, even for one entry.
9. Chunking is by entry count (1000) and uncompressed size (5MB), whichever binds first, and logs are encoded one at a time into the chunk under construction so that peak memory is one chunk, not one batch.
10. An entry that alone exceeds the body limit is dropped with one warning.
11. `202` is success. `401`/`403` abandon the rest of the stream's batch with one error. Any other status or transport error drops that chunk and continues. Nothing is retried.
12. One deadline for the whole stream batch (`datadogBatchTimeout`) plus one per request (`datadogRequestTimeout`), so a slow destination cannot outlive the drain lock.
13. `ResolveHostname` returns `""` for an unusable host; the syslog encoder turns that into `-`, the datadog encoder omits the field.
14. `datadogStatus` reads the same table as syslog severity, so an activity type is classified once.
15. Configured tags are sorted by key; `app_id` and `activity_type` are appended from the event and override a configured pair of the same name.

## 10. Atomic commit plan

### Commit 1 — `[Telemetry] Let an SSRF-safe client verify against a given certificate authority`

- `pkg/util/httputil/ext_client.go`, `pkg/util/httputil/ext_client_test.go`

Self-contained and used by nothing yet. A regression here would weaken every project-configured fetch in the product, so it is reviewed on its own rather than inside a feature commit.

### Commit 2 — `[Telemetry] Add the Datadog log object encoder`

- `pkg/lib/telemetry/auditlogstreaming/datadog.go`, `datadog_test.go`
- `hostname.go`, `encoder.go` and their tests, for the `""`/`-` split

Pure functions. Reviewable directly against [the log object](../../specs/audit-log-streaming.md#the-log-object), which is where a silent mistake would otherwise live — the intake answers `202` to a malformed body, so nothing downstream reports an encoder defect.

### Commit 3 — `[Telemetry] Resolve datadog streams and the http transport`

- `resolve.go`, `resolve_test.go`, and the mechanical `transport.go` adjustments for `*ResolvedTCP` / `*ResolvedSyslog`

Depends on part 01's config and secret types. The syslog path keeps working throughout; nothing calls the datadog fields yet.

### Commit 4 — `[Telemetry] Deliver datadog streams to the logs HTTP intake`

- `datadog_transport.go`, `datadog_transport_test.go`
- `transport.go` (the type switch, `sendSyslogTCP`, the new `SenderImpl` fields), `transport_test.go`
- `deps.go`, `pkg/lib/deps/deps_config.go`
- **regenerated `cmd/authgear/background/wire_gen.go`, in this commit** — the sender's dependencies change here, so the generated wiring moves with them or the tree does not build.

This is the commit where a datadog stream starts delivering.
