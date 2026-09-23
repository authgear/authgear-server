- [Audit Log Streaming](#audit-log-streaming)
  * [About Audit Log Streaming](#about-audit-log-streaming)
  * [Configuration](#configuration)
    + [The stream object](#the-stream-object)
    + [The syslog object](#the-syslog-object)
    + [The datadog object](#the-datadog-object)
    + [The tcp object](#the-tcp-object)
    + [The http object](#the-http-object)
    + [The tls secret](#the-tls-secret)
    + [The datadog secret](#the-datadog-secret)
    + [The feature config](#the-feature-config)
  * [Type: syslog](#type-syslog)
    + [`rfc5424`](#rfc5424)
      - [PRI](#pri)
      - [MSG](#msg)
    + [Framing](#framing)
  * [Type: datadog](#type-datadog)
    + [The request](#the-request)
    + [The log object](#the-log-object)
      - [Datadog standard attributes](#datadog-standard-attributes)
      - [`status`](#status)
      - [`ddtags`](#ddtags)
      - [`authgear.event`](#authgearevent)
    + [Example](#example)
  * [Delivery](#delivery)
  * [Use cases](#use-cases)
    + [UC1. Stream audit logs to an OpenTelemetry Collector](#uc1-stream-audit-logs-to-an-opentelemetry-collector)
    + [UC2. Authenticate to the collector with mutual TLS](#uc2-authenticate-to-the-collector-with-mutual-tls)
    + [UC3. Stream audit logs to two collectors](#uc3-stream-audit-logs-to-two-collectors)
    + [UC4. Stream audit logs to Datadog](#uc4-stream-audit-logs-to-datadog)
  * [Caveats](#caveats)
    + [Caveats of type: datadog](#caveats-of-type-datadog)
  * [Future works](#future-works)
    + [Additional types, transports and formats](#additional-types-transports-and-formats)
    + [Retry](#retry)
    + [Date range replay](#date-range-replay)
    + [Custom hook](#custom-hook)
  * [Appendix: researches](#appendix-researches)
    + [What SIEM services expect](#what-siem-services-expect)
    + [OpenTelemetry collectors](#opentelemetry-collectors)
    + [Datadog](#datadog)
    + [Observations](#observations)
    + [How the design fits](#how-the-design-fits)

# Audit Log Streaming

Audit Log Streaming delivers [audit log](./audit-log.md) entries to an external log collector shortly after they occur.

## About Audit Log Streaming

- It is configured per project in `authgear.yaml`.
- It streams the same entries that are written to the audit database. An entry is streamed if and only if it is persisted as an audit log.
- It is asynchronous. It never blocks and never fails an end-user request.
- It is batched. An entry is queued when it occurs, and the queue is delivered once a minute. An entry therefore reaches the collector up to a minute after it occurs.
- It is at-most-once. There is no acknowledgement from the collector.
- The audit database remains the source of truth. Streaming is an additional copy.

## Configuration

```yaml
telemetry:
  audit_logs:
    streams:
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
        tls:
          enabled: true
      syslog:
        format: rfc5424
        framing: newline
        facility: local0
        app_name: authgear
        structured_data_id: authgear
    - name: datadog
      type: datadog
      transport: http
      datadog:
        site: datadoghq.com
        service: authgear
        source: authgear
        tags:
          env: production
```

- `telemetry.audit_logs.streams` is a list. Every configured stream receives every audit log entry.
- When `telemetry.audit_logs.streams` is absent or empty, no streaming occurs.

### The stream object

| Field | Required | Values | Default | Description |
|---|---|---|---|---|
| `name` | yes | string, 1-63 chars, `[a-zA-Z0-9_-]` | | Identifies the stream in logs and metrics. Unique within the project. |
| `type` | yes | `syslog`, `datadog` | | The encoding of a record. Selects which encoding object is allowed. |
| `transport` | yes | `tcp`, `http` | | How records are delivered. Selects which transport object is allowed. |
| `syslog` | yes when `type` is `syslog` | object | | See [the syslog object](#the-syslog-object). |
| `datadog` | no, and only when `type` is `datadog` | object | `{}` | See [the datadog object](#the-datadog-object). |
| `tcp` | yes when `transport` is `tcp` | object | | See [the tcp object](#the-tcp-object). |
| `http` | no, and only when `transport` is `http` | object | `{}` | See [the http object](#the-http-object). |

`type` and `transport` are independent selectors. Only these combinations are valid. Any other combination is rejected when saving `authgear.yaml`.

| `type` | `transport` |
|---|---|
| `syslog` | `tcp` |
| `datadog` | `http` |

`datadog` is paired with `http` only. Datadog states that [TCP log collection is not supported and gives no delivery or reliability guarantee](https://docs.datadoghq.com/logs/log_collection/?tab=http#custom-log-forwarding), so a syslog stream is not a way to reach Datadog directly. See [Datadog](#datadog).

A new row is added here as [additional types and transports](#additional-types-transports-and-formats) are supported.

### The syslog object

| Field | Required | Values | Default | Description |
|---|---|---|---|---|
| `format` | yes | `rfc5424` | | The syslog dialect. |
| `framing` | yes | `octet_counting`, `newline` | | How records are delimited in the TCP stream. Must match the receiver. |
| `facility` | no | `user`, `local0` - `local7` | `local0` | The syslog facility. |
| `app_name` | no | string, 1-48 printable US-ASCII chars | `authgear` | The RFC 5424 APP-NAME. |
| `structured_data_id` | no | string, 1-32 printable US-ASCII chars, no `=`, `]`, `"`, space | `authgear` | The RFC 5424 SD-ID. |

`format` and `framing` are required because they change the bytes on the wire and must match the receiver configuration.

`facility` names the same set of facilities redis.conf's `syslog-facility` accepts: `user` or `local0` through `local7`. Each name maps to its numeric facility code from Table 1 in [RFC5424 section-6.2.1](https://datatracker.ietf.org/doc/html/rfc5424#section-6.2.1) — `user` is `1`, `local0`-`local7` are `16`-`23` — which is the code used in [PRI](#pri).

The default, `local0` (numeric code `16`, `local use 0` of that table), does not collide with the facilities a host uses for its own messages.

### The datadog object

Every field has a default, so the object itself is optional. The API key is not here, it is in [the datadog secret](#the-datadog-secret).

| Field | Required | Values | Default | Description |
|---|---|---|---|---|
| `site` | no | a Datadog site parameter, see below | `datadoghq.com` | The Datadog site of the organization. Determines the intake endpoint. |
| `service` | no | string, 1-100 chars | `authgear` | The Datadog `service` of every log. |
| `source` | no | string, 1-100 chars | `authgear` | The Datadog `ddsource` of every log. |
| `tags` | no | map of string to string, at most 20 pairs | `{}` | Added to `ddtags` of every log. See [`ddtags`](#ddtags). |

- `site` is the site parameter of one of the [Datadog sites](https://docs.datadoghq.com/getting_started/site/), not its display name -- for example `datadoghq.com`, `us3.datadoghq.com`, `us5.datadoghq.com`, `datadoghq.eu`, `ap1.datadoghq.com`, `ap2.datadoghq.com`, `uk1.datadoghq.com`, `ddog-gov.com`, `us2.ddog-gov.com`. Authgear does not validate it against that list: `site` is a free-form string, so which sites are usable is Datadog's deployment to govern, not this project's. A value Datadog does not recognise fails at delivery time, as an unreachable `http.endpoint` does, not at config save time.
- `site` is rejected when `http.endpoint` is set. They are two ways to say the same thing, so exactly one of them is used.
- `source` becomes `ddsource`, which is how Datadog selects the integration pipeline that post-processes a log. Datadog has no `authgear` integration, so no pipeline is installed for the default value. This is why Authgear maps the [Datadog standard attributes](#datadog-standard-attributes) itself instead of relying on a pipeline. Change `source` only when the project has its own pipeline keyed on another value.
- `service` and `source` are also tags in Datadog, so their values are subject to [Datadog's tag rules](https://docs.datadoghq.com/getting_started/tagging/#define-tags).

### The tcp object

| Field | Required | Values | Default | Description |
|---|---|---|---|---|
| `address` | yes | `host:port` | | The collector address. |
| `tls` | no | object | disabled | TLS settings. |
| `tls.enabled` | yes when `tls` is present | boolean | | When true, the connection is wrapped in TLS. |

Certificates are not configured here. They are in [the tls secret](#the-tls-secret), and their presence determines the behaviour.

- The negotiated TLS version is at least 1.2.
- The collector certificate is verified using the host of `address` as the expected name.
- When the stream has a `certificate_authority` in the tls secret, the collector certificate is verified against it only. Otherwise it is verified against the system trust store.
- When the stream has a `client_certificate` in the tls secret, Authgear presents it. This is mutual TLS.

### The http object

| Field | Required | Values | Default | Description |
|---|---|---|---|---|
| `endpoint` | no | absolute `http` or `https` URL | derived from the encoding object | The exact URL requests are sent to. |

- `endpoint` is rejected when `datadog.site` is set. Without it, the URL is derived from `datadog.site`, see [the request](#the-request).
- `endpoint` exists so that a project can put something of its own in front of the destination — a [Datadog Agent](https://docs.datadoghq.com/agent/), an [Observability Pipelines Worker](https://docs.datadoghq.com/observability_pipelines/), or a forward proxy. The body is unchanged, so whatever receives it has to accept the destination's payload.
- The scheme is `http` or `https`, the same two a [hook](./hook.md) URL accepts. Nothing else is.
- A plaintext `http` endpoint sends the `api_key` across the network in the clear, in a header, on every request. That is worse than it is for a hook, which carries a signature rather than a credential, so use `http` only for a destination on a network the deployment controls — an Agent on loopback, or a worker in the same cluster. See [the caveats](#caveats-of-type-datadog).
- For an `https` endpoint, the negotiated TLS version is at least 1.2, and the server certificate is verified using the host of the URL as the expected name.
- When the stream has a `certificate_authority` in [the tls secret](#the-tls-secret), the server certificate is verified against it only. Otherwise it is verified against the system trust store. This is for a self-hosted `https` endpoint with a private certificate authority. `client_certificate` is not used by the `http` transport.
- The request is subject to the deployment's fetch address policy, as an event webhook is. An `endpoint` that resolves to an address that is not publicly routable is refused, and the chunk is dropped with one warning. An operator who runs the destination inside their own network lifts that with `http.insecure_fetch_address_allowed`, or names the host in `http.insecure_fetch_address_allowed_hosts`, in `authgear.features.yaml`. Neither is writable through the portal or the Admin API, so a project cannot grant itself the deployment's internal network by configuring a stream.

### The tls secret

The certificate material is in `authgear.secrets.yaml` under the key `telemetry.audit_logs.streams.tls`. It is a list keyed by stream name.

```yaml
secrets:
- key: telemetry.audit_logs.streams.tls
  data:
  - stream_name: collector
    client_certificate:
      certificate:
        pem: |
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
      key:
        kty: RSA
        ...
    certificate_authority:
      pem: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
```

| Field | Required | Description |
|---|---|---|
| `stream_name` | yes | The `name` of the stream in `authgear.yaml` this material belongs to. |
| `client_certificate` | no | The certificate and private key Authgear presents to the collector. |
| `client_certificate.certificate` | yes when `client_certificate` is present | An `X509Certificate`. `pem` holds the certificate, followed by any intermediates. |
| `client_certificate.key` | yes when `client_certificate` is present | The private key of the certificate, as a `JWK`. |
| `certificate_authority` | no | An `X509Certificate`. `pem` holds the certificate authority certificate. |

- An item declares at least one of `client_certificate` and `certificate_authority`.
- `client_certificate` and `certificate_authority` are independent. Either may be declared without the other.
- An item whose `stream_name` matches no stream, or whose stream does not use TLS, is ignored. It is not an error, so removing a stream does not break the project.
- An item whose `stream_name` matches no stream is removed the next time `authgear.yaml` is saved. Material of a stream that still exists is kept, even while its `tcp.tls` is disabled, so that disabling TLS and re-enabling it does not require re-uploading the certificates.
- The certificate authority that signs the client certificate is chosen by the collector. It does not have to be the one in `certificate_authority`.
- Certificate expiry is not tracked. A stream fails to connect once its client certificate expires.

### The datadog secret

The Datadog API key is in `authgear.secrets.yaml` under the key `telemetry.audit_logs.streams.datadog`. It is a list keyed by stream name.

```yaml
secrets:
- key: telemetry.audit_logs.streams.datadog
  data:
  - stream_name: datadog
    api_key: "0123456789abcdef0123456789abcdef"
```

| Field | Required | Description |
|---|---|---|
| `stream_name` | yes | The `name` of the stream in `authgear.yaml` this key belongs to. |
| `api_key` | yes | A Datadog API key, from Organization Settings > API Keys. It is sent as the `DD-API-KEY` header. |

- It is an API key, not an application key. An application key does not authenticate to the logs intake.
- A Datadog API key is not scoped. It can write to every intake of the organization, so treat it as a full-organization write credential.
- A stream of `type: datadog` with no item in this secret is an invalid configuration, the way a SAML service provider with no certificates is. Saving such an `authgear.yaml` fails, and so does loading one: the pair is checked every time the configuration is read, not only when it is saved. Unlike the tls secret, which is optional material, the key is what makes the stream able to deliver at all.
- The consequence is that the secret is written before, or together with, the stream that needs it, and that the stream is removed before the key it uses. A project left with one and not the other does not load, which for a self-hosted deployment editing the files by hand means the project stops serving until the pair is complete again.
- An item whose `stream_name` matches no stream is ignored, and is removed the next time `authgear.yaml` is saved.
- Key rotation is manual. Authgear does not detect a revoked key other than by the `403` it gets from the intake, see [the request](#the-request).

### The feature config

Availability is gated in `authgear.features.yaml`.

```yaml
telemetry:
  audit_logs:
    streaming:
      disabled: false
```

| Field | Required | Values | Default | Description |
|---|---|---|---|---|
| `disabled` | no | boolean | `false` | When true, the project cannot configure `telemetry.audit_logs.streams`. |

- Saving an `authgear.yaml` with a non-empty `telemetry.audit_logs.streams` fails when `disabled` is true.
- A project whose `authgear.yaml` already has streams stops streaming when `disabled` becomes true. The configured streams are ignored at runtime.
- The gate is on streaming as a whole. A `type` is not gated on its own.

## Type: syslog

A stream of `type: syslog` sends one syslog message per entry. The dialect is `syslog.format`.

### `rfc5424`

A message is a single [RFC5424](https://datatracker.ietf.org/doc/html/rfc5424) message.

| Field | Value |
|---|---|
| PRI | See [PRI](#pri). |
| VERSION | `1` |
| TIMESTAMP | `context.timestamp` as RFC 3339 in UTC. The event carries whole seconds, so there is no fractional part. |
| HOSTNAME | The host of `http.public_origin`, without the port. It identifies the project, not the Authgear process that delivered the entry. `-` when it is not a valid HOSTNAME. |
| APP-NAME | `syslog.app_name`. |
| PROCID | `-` |
| MSGID | `authgear-audit-log` |
| STRUCTURED-DATA | One element, see below. |
| MSG | See [MSG](#msg). |

MSGID is the constant `authgear-audit-log`. It does not carry the activity type, because [RFC5424 section-6.2.7](https://datatracker.ietf.org/doc/html/rfc5424#section-6.2.7) limits MSGID to 32 characters and activity types are up to 60 characters.

The structured data element is:

```
[<structured_data_id> app_id="..." id="..." activity_type="..." user_id="..." client_id="..." ip_address="..."]
```

- The values come from the event. `id` is `id` and `activity_type` is `type`. `app_id`, `user_id`, `client_id` and `ip_address` are of `context`.
- A parameter is omitted when its value is empty.
- `"`, `\` and `]` in a parameter value are escaped as `\"`, `\\` and `\]`.

#### PRI

PRI is `facility * 8 + severity`, where `facility` is the numeric code `syslog.facility` maps to (see [the syslog object](#the-syslog-object)).

Severity is `4` (warning) for these activity types:

- `email.error`
- `sms.error`
- `whatsapp.error`

Severity is `6` (info) for every other activity type.

The severity values are the severity codes of Table 2 in [RFC5424 section-6.2.1](https://datatracker.ietf.org/doc/html/rfc5424#section-6.2.1).

- The list is explicit. A new activity type is `6` (info) until it is added to the list.
- Severity is a convenience for receiver-side filtering. `activity_type` is the authoritative classification of an entry.

#### MSG

MSG is the [event](./event.md) object as a single-line UTF-8 JSON object. It is the same object the portal shows as the Raw Event Log of an audit log entry.

```json
{
  "id": "00000000000a5a60",
  "seq": 678496,
  "type": "user.authenticated",
  "payload": {
    "session": { ... },
    "user": { ... }
  },
  "context": {
    "app_id": "myproject",
    "user_id": "00000000-0000-0000-0000-000000000001",
    "client_id": "0000000000000000",
    "ip_address": "203.0.113.9",
    "user_agent": "Mozilla/5.0 ...",
    "geo_location_code": "TW",
    "language": "en",
    "preferred_languages": ["en", "zh-HK", "zh"],
    "timestamp": 1785148956,
    "triggered_by": "user",
    "tracking_id": "...",
    "audit_context": { ... },
    "oauth": { "state": "..." }
  }
}
```

- `payload` is specific to the event type.
- `context` carries the request attributes. A field of it is omitted when its value is empty.
- MSG is not prefixed with a BOM.
- A message is commonly a few kilobytes. A `user.authenticated` payload embeds the whole user and session object.

### Framing

| `framing` | Frame |
|---|---|
| `octet_counting` | `<MSG-LEN> <SYSLOG-MSG>`, where `MSG-LEN` is the byte length of `SYSLOG-MSG`. Octet counting of [RFC6587 section-3.4.1](https://datatracker.ietf.org/doc/html/rfc6587#section-3.4.1). |
| `newline` | `<SYSLOG-MSG>\n`. Non-transparent framing of [RFC6587 section-3.4.2](https://datatracker.ietf.org/doc/html/rfc6587#section-3.4.2), with an LF trailer. |

`newline` is safe for every entry. MSG is the output of a JSON encoder, so an LF inside a value is escaped as `\n`, and no other field of the message can contain an LF.

The two framings are subject to different size limits on the receiver. In the OpenTelemetry syslog receiver:

| `framing` | Receiver limit | Default |
|---|---|---|
| `octet_counting` | `max_octets` | 8192 bytes |
| `newline` | the TCP `max_log_size` | 1 MiB |

An entry whose `payload` is large can exceed 8192 bytes. Use `octet_counting` only when the receiver requires it, and confirm it allows `max_octets` to be raised.

## Type: datadog

A stream of `type: datadog` sends entries to the Datadog logs HTTP intake. One entry becomes one Datadog log.

The intake is [Send logs](https://docs.datadoghq.com/api/latest/logs/send-logs/). This section states what Authgear sends and how an entry maps onto a Datadog log. It does not restate the intake's own contract — its payload fields, limits and status codes are in that document.

### The request

A chunk of a batch is one `POST` to `https://http-intake.logs.<site>/api/v2/logs`, where `<site>` is `datadog.site`. When `http.endpoint` is set, that URL is used verbatim instead.

- The `DD-API-KEY` header carries the `api_key` of [the datadog secret](#the-datadog-secret).
- The body is a gzipped JSON array of [log objects](#the-log-object), even when it holds a single entry. The intake also accepts a bare object, but an array keeps one code path.
- A batch is split into as many requests as the intake's limits on entries per request and on uncompressed body size require. The chunks are sent one after another, in the order the entries occurred.
- A single entry that alone exceeds the body limit is dropped and one warning is logged. It cannot be delivered.
- `202` is success. Any other status, and any transport error, drops that chunk and logs one warning with the status and the response body. The remaining chunks of the batch are still sent, and nothing is retried. See [retry](#retry).
- `202` means accepted, not stored as intended. The intake answers `202` with an empty body, and it answered `202` to a deliberately malformed body when this was tested against a live account. A defect in the encoder is therefore silent — it is not reported by the status code, and it is not visible until someone looks in Datadog.
- `401` and `403` mean the `api_key` is missing, wrong or revoked. That is a configuration error, and it drops every chunk until the key is fixed.

### The log object

An entry becomes one JSON object. The [event](./event.md) object is carried verbatim under `authgear.event`, and every other attribute is a projection of it, so nothing is lost and nothing has to be reconstructed on the Datadog side.

| Attribute | Source | Why this attribute |
|---|---|---|
| `ddsource` | `datadog.source` | A [request body field](https://docs.datadoghq.com/api/latest/logs/send-logs/), "the integration name associated with your log: the technology from which the log originated". Datadog stores it as the [`source` reserved attribute](https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#reserved-attributes), and selects the [integration pipeline](https://docs.datadoghq.com/logs/log_configuration/pipelines/) by it. |
| `service` | `datadog.service` | A [request body field](https://docs.datadoghq.com/api/latest/logs/send-logs/) that is also the [`service` reserved attribute](https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#reserved-attributes), "the name of the application or service generating the log events". |
| `hostname` | The host of `http.public_origin`, without the port. Omitted when it is not a valid host. | A [request body field](https://docs.datadoghq.com/api/latest/logs/send-logs/), "the name of the originating host of the log", which Datadog stores as the [`host` reserved attribute](https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#reserved-attributes). A project is the closest thing a multi-tenant Authgear has to a host, so the value identifies the project, not the Authgear process that delivered the entry. |
| `ddtags` | `datadog.tags`, plus `app_id` and `activity_type` | A [request body field](https://docs.datadoghq.com/api/latest/logs/send-logs/), "tags associated with your logs". Tags are what Datadog index filters, exclusion filters and retention filters are written against. See [`ddtags`](#ddtags). |
| `message` | `type` | A [request body field](https://docs.datadoghq.com/api/latest/logs/send-logs/) that is also the [`message` reserved attribute](https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#reserved-attributes): Datadog "ingests the value of the `message` attribute as the body of the log entry". It is the line the Log Explorer shows and the text free-text search matches, so it holds the one field that identifies an entry at a glance. |
| `status` | Derived from `type` | Not a request body field. It is the [`status` reserved attribute](https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#reserved-attributes), the severity of a log, which Datadog uses to define patterns and to colour the Log Explorer. See [`status`](#status). |
| `timestamp` | `context.timestamp`, in milliseconds | Not a request body field. It is the date of the log. Without it Datadog stamps the log with its ingestion time, which is up to a minute after the entry occurred. |
| `evt.name` | `type` | A [Datadog standard attribute](#datadog-standard-attributes). |
| `usr.id` | `context.user_id` | A [Datadog standard attribute](#datadog-standard-attributes). |
| `network.client.ip` | `context.ip_address` | A [Datadog standard attribute](#datadog-standard-attributes). |
| `network.client.geoip.country.iso_code` | `context.geo_location_code` | A [Datadog standard attribute](#datadog-standard-attributes). |
| `http.useragent` | `context.user_agent` | A [Datadog standard attribute](#datadog-standard-attributes). |
| `http.url` | `context.audit_context.http_url` | A [Datadog standard attribute](#datadog-standard-attributes). |
| `http.referer` | `context.audit_context.http_referer` | A [Datadog standard attribute](#datadog-standard-attributes). |
| `authgear.event` | The whole event object | Nothing Datadog defines can carry it, and it is what makes the log lossless. See [`authgear.event`](#authgearevent). |

- `ddsource`, `ddtags`, `hostname`, `message` and `service` are [request body fields](https://docs.datadoghq.com/api/latest/logs/send-logs/), which Datadog stores as `source`, the log's tags, `host`, `message` and `service`. Every other key is a custom attribute.
- `status` and `timestamp` are custom attributes that [preprocessing for JSON logs](https://docs.datadoghq.com/logs/log_configuration/pipelines/?tab=date#preprocessing-for-json-logs) promotes to the log's status and date. The event carries whole seconds, so `timestamp` always ends in `000`.
- An attribute is omitted when its source value is empty. `ddsource`, `service`, `message`, `status`, `timestamp` and `authgear.event` are always present.
- Nothing else is written at the top level.

#### Datadog standard attributes

These are [Datadog standard attributes](https://docs.datadoghq.com/standard-attributes/), not the Authgear [standard attributes](./user-profile/design.md#standard-attributes) of a user profile.

| Datadog standard attribute | Value | Why |
|---|---|---|
| `evt.name` | `type` | The Event Name facet, which an audit log stream is grouped and filtered by. |
| `usr.id` | `context.user_id` | User attribution. |
| `network.client.ip` | `context.ip_address` | |
| `network.client.geoip.country.iso_code` | `context.geo_location_code` | |
| `http.useragent` | `context.user_agent` | |
| `http.url` | `context.audit_context.http_url` | |
| `http.referer` | `context.audit_context.http_referer` | Set for entries from the portal and the Site Admin API, absent otherwise. |

- `context.audit_context` is a free-form object. An attribute is omitted when its key is absent or is not a string.
- `network.client.geoip.*` beyond `country.iso_code`, `http.useragent_details.*` and `http.url_details.*` are not sent. The [GeoIP](https://docs.datadoghq.com/logs/log_configuration/processors/geoip_parser/), [User-Agent](https://docs.datadoghq.com/logs/log_configuration/processors/user_agent_parser/) and [URL](https://docs.datadoghq.com/logs/log_configuration/processors/url_parser/) parsers that produce them need a pipeline, and none is installed for `ddsource: authgear`. The values are sent on the paths those parsers read, so adding one is a one-click change; adding the GeoIP parser then replaces the country code with Datadog's own.
- `usr.name` and `usr.email` are not sent. The event context carries only `user_id`, and a payload with an email address does not always mean the acting user's — `email.sent` has a sender, a recipient and no user. Those values still reach Datadog under [`authgear.event`](#authgearevent).
- `evt.outcome` is not sent. An activity type does not carry an outcome in a way that generalises: some encode it in the name, such as `authentication.primary.password.failed`, and some do not.

#### `status`

`status` is `warning` for these activity types:

- `email.error`
- `sms.error`
- `whatsapp.error`

`status` is `info` for every other activity type.

This is the same list that [PRI](#pri) uses, so the two types classify an entry the same way. `warning` and `info` are syslog severity names, which is what Datadog's status remapper expects.

- The list is explicit. A new activity type is `info` until it is added to the list.
- `status` is a convenience for filtering and for the Datadog UI. `evt.name` is the authoritative classification of an entry.

#### `ddtags`

`ddtags` is [the intake's tag field](https://docs.datadoghq.com/api/latest/logs/send-logs/), a comma-separated string of `key:value` tags.

```
env:production,team:security,app_id:myproject,activity_type:user.authenticated
```

It is the pairs of `datadog.tags`, followed by:

| Tag | Source |
|---|---|
| `app_id` | `context.app_id` |
| `activity_type` | `type` |

- `app_id` and `activity_type` are always appended. A pair of `datadog.tags` with either key is dropped in favour of the Authgear value.
- Authgear does not rewrite a tag to fit [Datadog's tag rules](https://docs.datadoghq.com/getting_started/tagging/#define-tags). A tag that does not conform to them is normalised or rejected by Datadog.
- `app_id` and `activity_type` are repeated here although they are also attributes. A tag is what Datadog index filters, exclusion filters and retention filters are written against, so the two most likely filtering keys are available as tags.
- `env` is not set by Authgear. A project that uses [unified service tagging](https://docs.datadoghq.com/getting_started/tagging/unified_service_tagging/) puts `env` in `datadog.tags`.

#### `authgear.event`

`authgear.event` is the [event](./event.md) object, unchanged. It is the same object [MSG](#msg) carries and the same object the portal shows as the Raw Event Log of an audit log entry.

```json
{
  "authgear": {
    "event": {
      "id": "00000000000a5a60",
      "seq": 678496,
      "type": "user.authenticated",
      "payload": { ... },
      "context": { ... }
    }
  }
}
```

- The outer key is the vendor and the inner key the record kind. A Datadog facet is defined per attribute path for the whole organization, so a generic path such as `audit_log` collides with other log sources, and `event` would sit beside Datadog's own `evt` namespace. `authgear` is also what `structured_data_id` defaults to in [the syslog type](#rfc5424).
- Its fields are searchable as `@authgear.event.id`, `@authgear.event.context.client_id` and so on, with no pipeline and [no facet](https://docs.datadoghq.com/logs/explorer/facets/).
- [Free text search](https://docs.datadoghq.com/logs/explorer/search_syntax/) does not reach them. It covers only `message`, which is the activity type. Use `*:<value>` or an attribute search.
- Datadog's per-log attribute limits apply, see [the caveats](#caveats-of-type-datadog).

### Example

A `user.authenticated` entry becomes one element of the array:

```json
{
  "ddsource": "authgear",
  "service": "authgear",
  "hostname": "myproject.authgear.cloud",
  "ddtags": "env:production,app_id:myproject,activity_type:user.authenticated",
  "message": "user.authenticated",
  "status": "info",
  "timestamp": 1785148956000,
  "evt": {
    "name": "user.authenticated"
  },
  "usr": {
    "id": "00000000-0000-0000-0000-000000000001"
  },
  "network": {
    "client": {
      "ip": "203.0.113.9",
      "geoip": { "country": { "iso_code": "TW" } }
    }
  },
  "http": {
    "useragent": "Mozilla/5.0 ...",
    "url": "https://myproject.authgear.cloud/oauth2/token",
    "referer": "https://myproject.authgear.cloud/login"
  },
  "authgear": {
    "event": {
      "id": "00000000000a5a60",
      "seq": 678496,
      "type": "user.authenticated",
      "payload": {
        "session": { ... },
        "user": { ... }
      },
      "context": {
        "app_id": "myproject",
        "user_id": "00000000-0000-0000-0000-000000000001",
        "client_id": "0000000000000000",
        "ip_address": "203.0.113.9",
        "user_agent": "Mozilla/5.0 ...",
        "geo_location_code": "TW",
        "language": "en",
        "preferred_languages": ["en", "zh-HK", "zh"],
        "timestamp": 1785148956,
        "triggered_by": "user",
        "tracking_id": "...",
        "audit_context": { ... },
        "oauth": { "state": "..." }
      }
    }
  }
}
```

`payload` and `context` are abbreviated here, see [`authgear.event`](#authgearevent).

## Delivery

An entry is queued when it is persisted. Once a minute the queue of a project is taken as a batch and sent to every configured stream of that project.

- Every entry is sent to the destination of every configured stream once. A batch that fails to send is not retried, and its entries are dropped. Delivery is therefore at-most-once.
- A dropped entry stays in the audit database, and is retrievable with the Admin API `auditLogs` query for the retention period.
- Entries are delivered in the order they occurred.
- The queue is bounded. A project queues at most 10000 entries between two deliveries; beyond that the oldest queued entry is dropped to make room. The bound is reached only when a project produces more than 10000 entries in a minute, or when delivery has stopped running.
- Taking the batch clears the queue, whether or not the send succeeds. Nothing accumulates across deliveries.
- A batch is one unit for the `tcp` transport. For the `http` transport it is split into as many requests as the destination's limits require, see [the request](#the-request).

## Use cases

### UC1. Stream audit logs to an OpenTelemetry Collector

`authgear.yaml`:

```yaml
telemetry:
  audit_logs:
    streams:
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
      syslog:
        format: rfc5424
        framing: newline
```

The syslog receiver of the collector must be configured to match:

```yaml
receivers:
  syslog:
    protocol: rfc5424
    tcp:
      listen_address: 0.0.0.0:5140
```

`protocol` has to be stated, and it has to be `rfc5424`. `framing: newline` needs nothing further, because `enable_octet_counting` is `false` by default and the receiver then reads one message per line.

A `user.authenticated` entry is sent as:

```
<134>1 2026-07-27T10:42:36Z myproject.authgear.cloud authgear - authgear-audit-log [authgear app_id="myproject" id="00000000000a5a60" activity_type="user.authenticated" user_id="00000000-0000-0000-0000-000000000001" client_id="0000000000000000" ip_address="203.0.113.9"] {"id":"00000000000a5a60","seq":678496,"type":"user.authenticated","payload":{ ... },"context":{ ... }}
```

terminated by an LF. MSG is abbreviated here, see [MSG](#msg).

To recover MSG as fields, add a JSON parser on the message body in the collector.

### UC2. Authenticate to the collector with mutual TLS

The collector requires every sender to present a client certificate. `tcp.tls.enabled` turns on TLS. The client certificate in `authgear.secrets.yaml` is what makes the connection mutually authenticated.

`authgear.yaml`:

```yaml
telemetry:
  audit_logs:
    streams:
    - name: collector
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:6514
        tls:
          enabled: true
      syslog:
        format: rfc5424
        framing: newline
```

`authgear.secrets.yaml`:

```yaml
secrets:
- key: telemetry.audit_logs.streams.tls
  data:
  - stream_name: collector
    client_certificate:
      certificate:
        pem: |
          -----BEGIN CERTIFICATE-----
          ...
          -----END CERTIFICATE-----
      key:
        kty: RSA
        ...
    certificate_authority:
      pem: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
```

The syslog receiver of the collector must be configured to match:

```yaml
receivers:
  syslog:
    protocol: rfc5424
    tcp:
      listen_address: 0.0.0.0:6514
      tls:
        cert_file: /path/to/collector.crt
        key_file: /path/to/collector.key
        client_ca_file: /path/to/client-ca.crt
```

- `cert_file` and `key_file` are the collector certificate and key, signed by the authority in `certificate_authority`.
- `client_ca_file` is the authority that signed `client_certificate`. Setting it is what makes the receiver require and verify a client certificate.

The connection fails when the collector does not present a certificate signed by the authority in `certificate_authority`, or when Authgear's client certificate is not signed by the authority in `client_ca_file`.

### UC3. Stream audit logs to two collectors

```yaml
telemetry:
  audit_logs:
    streams:
    - name: prod
      type: syslog
      transport: tcp
      tcp:
        address: collector.internal:5140
        tls:
          enabled: true
      syslog:
        format: rfc5424
        framing: newline
    - name: staging
      type: syslog
      transport: tcp
      tcp:
        address: staging-collector.internal:5140
      syslog:
        format: rfc5424
        framing: octet_counting
```

Both streams receive every entry. They are independent. A failure of one does not affect the other.

### UC4. Stream audit logs to Datadog

The project is on US1, the default site, and every field of the datadog object has a default. The configuration is the stream and the API key.

`authgear.yaml`:

```yaml
telemetry:
  audit_logs:
    streams:
    - name: datadog
      type: datadog
      transport: http
```

`authgear.secrets.yaml`:

```yaml
secrets:
- key: telemetry.audit_logs.streams.datadog
  data:
  - stream_name: datadog
    api_key: "0123456789abcdef0123456789abcdef"
```

Nothing is configured on the Datadog side. The logs arrive under `source:authgear service:authgear`, with the [Datadog standard attributes](#datadog-standard-attributes) already mapped.

A project on another site, or one that tags by environment, states it in the `datadog` object:

```yaml
      datadog:
        site: datadoghq.eu
        tags:
          env: production
```

A project that sends through a [Datadog Agent](https://docs.datadoghq.com/agent/) or an [Observability Pipelines Worker](https://docs.datadoghq.com/observability_pipelines/) of its own replaces the site with that URL. The body is unchanged, so the worker receives exactly what the intake would have.

```yaml
      http:
        endpoint: https://opw.internal:8282/api/v2/logs
```

A worker on the deployment's own network is reached only once the operator allows it, because the fetch address policy refuses a non-publicly-routable address by default. In `authgear.features.yaml`:

```yaml
http:
  insecure_fetch_address_allowed_hosts:
  - opw.internal
```

A destination that does not terminate TLS, such as an Agent listening on loopback, is reached over `http`. The `api_key` is then in the clear on that hop, which is the trade to make knowingly and only on a network the deployment controls.

```yaml
      http:
        endpoint: http://localhost:8126/api/v2/logs
```

## Caveats

- A long collector outage leaves a gap in the stream, which is not filled by any later delivery.
- Delivery runs in the Authgear background worker. A deployment that does not run it queues entries and delivers none of them.
- An entry that is queued but not yet delivered is lost if the queue store is lost. The entry remains in the audit database.
- `tcp.address` is not validated against private or link-local ranges, and the `tcp` transport connects to whatever it names.
- `http.endpoint` is not validated against those ranges either, but the `http` transport refuses to connect to one at delivery time unless the deployment allows it. This is the existing treatment of hook URLs, see [the http object](#the-http-object).
- The default `structured_data_id` of `authgear` is not of the form `name@<private-enterprise-number>` that [RFC5424 section-6.3.2](https://datatracker.ietf.org/doc/html/rfc5424#section-6.3.2) requires for non-IANA-registered SD-IDs. Set `structured_data_id` to a compliant value where the receiver enforces it.
- MSG is not prefixed with a BOM, which RFC 5424 recommends for UTF-8 content.

### Caveats of type: datadog

- Datadog rejects a log whose date is more than 18 hours in the past. A batch is delivered within a minute, so this is reached only after the background worker has been stopped for longer than that, and those entries are lost. It also bounds [date range replay](#date-range-replay).
- Datadog silently degrades a log that exceeds its per-log size or attribute limits: it truncates and still accepts the request. An entry is a few kilobytes today, well under them.
- The event payload carries personal data — email addresses, phone numbers, names, IP addresses — and is sent as-is, as is the query string of `http.url`. Redact with [Sensitive Data Scanner](https://docs.datadoghq.com/sensitive_data_scanner/), an exclusion filter, or a worker in front as in [UC4](#uc4-stream-audit-logs-to-datadog).
- Datadog bills by ingested volume and by indexed event. Authgear does not filter, so a project pays for every entry.
- A Datadog API key is not scoped to logs, and Authgear does not rotate it. Treat `api_key` as a full-organization write credential.
- `datadog.site` is not validated against Datadog's list of sites. A typo or a decommissioned site is not caught at save time -- it produces a URL that fails at every delivery, the same failure mode as a wrong `http.endpoint`.
- A `type: datadog` stream and its API key are validated as a pair on every configuration load, so deleting the key of a stream that still exists takes the whole project down, not just its streaming. See [the datadog secret](#the-datadog-secret).
- An `http.endpoint` with the `http` scheme puts the `api_key` on the wire in cleartext. Authgear accepts it, as it accepts a plaintext hook URL, and does not warn beyond this: the intended destination is a local Agent or an in-cluster worker. A plaintext endpoint on a public network exposes a full-organization write credential to anyone on the path.

## Future works

### Additional types, transports and formats

A new encoding is a new `type`. A new delivery method is a new `transport`. Possible additions:

| `type` | `format` | `transport` |
|---|---|---|
| `syslog` | `rfc3164` | `udp` |
| `otlp` | `protobuf`, `json` | `grpc`, `http` |
| `cef`, `leef` | | `tcp`, `udp` |

### Retry

A rejected chunk is dropped. Datadog documents which statuses are retryable, and a bounded retry with backoff would turn a short Datadog incident from a gap into a delay. It changes delivery from at-most-once to at-least-once for a chunk that is retried after a response was lost, so it also needs the destination to tolerate a duplicate. `authgear.event.id` is what a destination deduplicates on.

### Date range replay

An Admin API mutation that replays a date range of audit log entries to a stream, to fill a gap or to load history when a stream is first configured.

A replayed entry carries the same `id` as the original, which the destination can use to detect a duplicate. Neither syslog nor a collector deduplicates.

For a Datadog stream the replayable range is bounded by Datadog, which accepts a timestamp at most 18 hours in the past. Anything older arrives stamped with the replay time.

### Custom hook

A `type` whose encoding is supplied by the project. The hook receives an entry and returns the message to send, so that a project can adapt a stream to a destination without a change in Authgear. One use is adding the authentication token that destinations such as Sumo Logic Cloud Syslog and Loggly require in place of mutual TLS.

Tentative config:

```yaml
telemetry:
  audit_logs:
    streams:
    - name: sumologic
      type: hook
      transport: tcp
      tcp:
        address: syslog.collection.us2.sumologic.com:6514
        tls:
          enabled: true
      hook:
        url: authgeardeno:///deno/audit_stream.ts
        framing: newline
```

- The hook returns the complete message. Authgear does not check it.
- The hook runs for every entry, so a webhook is impractical. It would add one request per entry.

## Appendix: researches

### What SIEM services expect

| Service | Expected integration | How syslog is received |
|---|---|---|
| Splunk | HTTPS to HEC, concatenated JSON objects, each record nested under `event`. Syslog through SC4S. | [Splunk Connect for Syslog](https://splunk.github.io/splunk-connect-for-syslog/main/) |
| Microsoft Sentinel | HTTPS to the Logs Ingestion API, a JSON array matching a Data Collection Rule schema. Syslog or CEF through the Azure Monitor Agent. | [Forward syslog with the Azure Monitor Agent](https://learn.microsoft.com/en-us/azure/sentinel/forward-syslog-monitor-agent) |
| IBM QRadar | Syslog. LEEF, or RFC 5424. | [Sending syslog data to QRadar over TCP](https://www.ibm.com/docs/en/qradar-common?topic=cases-sending-syslog-data-qradar-over-tcp) |
| Elastic Security | HTTPS `_bulk`, NDJSON. Syslog through Elastic Agent. | [Custom TCP Logs integration](https://www.elastic.co/docs/reference/integrations/tcp) |
| Google SecOps | HTTPS ingestion API, UDM. Syslog through the Bindplane agent. | [Deploy the Bindplane agent for collection](https://docs.cloud.google.com/chronicle/docs/ingestion/use-bindplane-agent) |
| Datadog | HTTPS to the logs intake, a JSON array of log objects. | Only through the Datadog Agent or an Observability Pipelines Worker. [TCP log collection is not supported](https://docs.datadoghq.com/logs/log_collection/?tab=http#custom-log-forwarding). |

### OpenTelemetry collectors

A collector is not a destination. It accepts syslog inside the customer network and forwards to one or more destinations. Bindplane is a distribution of it, and its Syslog source is the [syslog receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/syslogreceiver/README.md) of the OpenTelemetry Collector.

- It accepts any RFC 5424 message. The dialect and the framing have to match the sender.
- The structured data parameters become fields of the log record with no configuration.
- MSG stays a string. A `json_parser` is configured on it to turn the event object into fields.

### Datadog

Datadog is the first destination reached without a collector, so it was researched on its own. The intake's own contract is in [Send logs](https://docs.datadoghq.com/api/latest/logs/send-logs/). These are the findings that shaped the design.

- **Datadog has no supported syslog endpoint of its own.** It has a legacy TCP intake that accepts raw, syslog or JSON lines, but its own documentation says [TCP log collection is not supported, that it gives no delivery or reliability guarantees, and that log data may be lost without notice](https://docs.datadoghq.com/logs/log_collection/?tab=http#custom-log-forwarding). Syslog reaches Datadog only through a customer-run [Agent](https://docs.datadoghq.com/agent/logs/?tab=tcpudp) or [Observability Pipelines Worker](https://docs.datadoghq.com/observability_pipelines/sources/syslog/), as it does for every other service in [the table](#what-siem-services-expect). An HTTPS type is what removes that collector from the deployment.
- **A JSON log is interpreted before any pipeline runs.** [Preprocessing for JSON logs](https://docs.datadoghq.com/logs/log_configuration/pipelines/?tab=date#preprocessing-for-json-logs) looks for the source, host, date, message, status and service among a fixed list of candidate attribute names. [The log object](#the-log-object) is named to hit those candidates, so a project configures nothing in Datadog.
- **A pipeline is selected by `ddsource`.** Datadog has no `authgear` integration, so no pipeline exists to map Authgear fields onto [Datadog standard attributes](#datadog-standard-attributes). Authgear maps them at the source instead. This is the one thing Datadog's [Auth0 integration](https://docs.datadoghq.com/integrations/auth0/) gets for free that Authgear does not.
- **The intake is not rate limited.** [Rate limits](https://docs.datadoghq.com/api/latest/rate-limits/) states that the API for sending logs is not rate limited, so chunks are not paced.
- **Precedent.** Auth0's Datadog log stream asks the tenant for two things, a Datadog region and a Datadog API key, and the logs arrive with `source` and `service` set to `auth0`. [UC4](#uc4-stream-audit-logs-to-datadog) asks for the same two things.

### Observations

- The HTTPS APIs have no common format. They differ in both record grouping and envelope, so an HTTPS integration serves one service per implementation.
- Every service accepts syslog, usually through a collector that runs inside the customer network and forwards to the service over HTTPS. Google SecOps recommends the Bindplane agent, a distribution of the OpenTelemetry Collector, for this. Datadog is no exception: its own TCP endpoint is unsupported, so syslog reaches it through a Datadog Agent or an Observability Pipelines Worker.
- CEF and LEEF are the only schemas parsed out of the box by more than one service. Both are flat `key=value`, so carrying the event object requires flattening it into dotted keys. Neither is supported at the moment. Either would be an additional `type`.
- A per-service HTTPS type is what removes the collector from the deployment. `type: syslog` asks the project to run one; `type: datadog` does not.

### How the design fits

- `type: syslog` reaches every service in the table through that service's collector, with no per-service implementation. `type: datadog` reaches Datadog with no collector at all, which is the only way to reach it directly.
- `type` names the encoding and `transport` names the delivery, and they stay independent. `datadog` is an encoding — a Datadog log object and the intake envelope around it — and `http` is a delivery. Naming the type after the transport instead, as in a single `http` type with a `format` field, would collapse the two axes, because every HTTPS destination differs in envelope, credential and endpoint, not only in record format.
- An entry is JSON in both types, and it is the same object the portal shows as the Raw Event Log — MSG for `syslog`, `authgear.event` for `datadog`. The routable metadata is repeated beside it, as structured data for `syslog` and as Datadog reserved and standard attributes for `datadog`, so a pipeline that does not parse the event object can still filter on `app_id`, `activity_type`, `user_id`, `client_id` and `ip_address`.
- Configuration is per project, as with Auth0 Log Streams and Okta Log Streaming. A Datadog stream asks for the same two things Auth0 asks for, a site and an API key, and both have a usable default or a secret of their own.
- Streaming carries only what is happening now. A gap is recoverable from the audit database, which retains every entry.
