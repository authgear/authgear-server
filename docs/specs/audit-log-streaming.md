- [Audit Log Streaming](#audit-log-streaming)
  * [About Audit Log Streaming](#about-audit-log-streaming)
  * [Configuration](#configuration)
    + [The stream object](#the-stream-object)
    + [The syslog object](#the-syslog-object)
    + [The tcp object](#the-tcp-object)
    + [The tls secret](#the-tls-secret)
    + [The feature config](#the-feature-config)
  * [Type: syslog](#type-syslog)
    + [`rfc5424`](#rfc5424)
      - [PRI](#pri)
      - [MSG](#msg)
    + [Framing](#framing)
  * [Delivery](#delivery)
  * [Use cases](#use-cases)
    + [UC1. Stream audit logs to an OpenTelemetry Collector](#uc1-stream-audit-logs-to-an-opentelemetry-collector)
    + [UC2. Authenticate to the collector with mutual TLS](#uc2-authenticate-to-the-collector-with-mutual-tls)
    + [UC3. Stream audit logs to two collectors](#uc3-stream-audit-logs-to-two-collectors)
  * [Caveats](#caveats)
  * [Future works](#future-works)
    + [Additional types, transports and formats](#additional-types-transports-and-formats)
    + [Date range replay](#date-range-replay)
    + [Custom hook](#custom-hook)
  * [Appendix: researches](#appendix-researches)
    + [What SIEM services expect](#what-siem-services-expect)
    + [OpenTelemetry collectors](#opentelemetry-collectors)
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
```

- `telemetry.audit_logs.streams` is a list. Every configured stream receives every audit log entry.
- When `telemetry.audit_logs.streams` is absent or empty, no streaming occurs.

### The stream object

| Field | Required | Values | Default | Description |
|---|---|---|---|---|
| `name` | yes | string, 1-63 chars, `[a-zA-Z0-9_-]` | | Identifies the stream in logs and metrics. Unique within the project. |
| `type` | yes | `syslog` | | The encoding of a record. Selects which encoding object is required. |
| `transport` | yes | `tcp` | | How records are delivered. Selects which transport object is required. |
| `tcp` | yes when `transport` is `tcp` | object | | See [the tcp object](#the-tcp-object). |
| `syslog` | yes when `type` is `syslog` | object | | See [the syslog object](#the-syslog-object). |

`type` and `transport` are independent selectors. Only these combinations are valid. Any other combination is rejected when saving `authgear.yaml`.

| `type` | `transport` |
|---|---|
| `syslog` | `tcp` |

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
- An item whose `stream_name` matches no stream, or whose stream does not enable `tcp.tls`, is ignored. It is not an error, so removing a stream does not break the project.
- An item whose `stream_name` matches no stream is removed the next time `authgear.yaml` is saved. Material of a stream that still exists is kept, even while its `tcp.tls` is disabled, so that disabling TLS and re-enabling it does not require re-uploading the certificates.
- The certificate authority that signs the client certificate is chosen by the collector. It does not have to be the one in `certificate_authority`.
- Certificate expiry is not tracked. A stream fails to connect once its client certificate expires.

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

## Delivery

An entry is queued when it is persisted. Once a minute the queue of a project is taken as a batch and sent to every configured stream of that project.

- Every entry is sent to the destination of every configured stream once. A batch that fails to send is not retried, and its entries are dropped. Delivery is therefore at-most-once.
- A dropped entry stays in the audit database, and is retrievable with the Admin API `auditLogs` query for the retention period.
- Entries are delivered in the order they occurred.
- The queue is bounded. A project queues at most 10000 entries between two deliveries; beyond that the oldest queued entry is dropped to make room. The bound is reached only when a project produces more than 10000 entries in a minute, or when delivery has stopped running.
- Taking the batch clears the queue, whether or not the send succeeds. Nothing accumulates across deliveries.

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

## Caveats

- A long collector outage leaves a gap in the stream, which is not filled by any later delivery.
- Delivery runs in the Authgear background worker. A deployment that does not run it queues entries and delivers none of them.
- An entry that is queued but not yet delivered is lost if the queue store is lost. The entry remains in the audit database.
- `tcp.address` is not validated against private or link-local ranges. This matches the existing treatment of hook URLs.
- The default `structured_data_id` of `authgear` is not of the form `name@<private-enterprise-number>` that [RFC5424 section-6.3.2](https://datatracker.ietf.org/doc/html/rfc5424#section-6.3.2) requires for non-IANA-registered SD-IDs. Set `structured_data_id` to a compliant value where the receiver enforces it.
- MSG is not prefixed with a BOM, which RFC 5424 recommends for UTF-8 content.

## Future works

### Additional types, transports and formats

A new encoding is a new `type`. A new delivery method is a new `transport`. Possible additions:

| `type` | `format` | `transport` |
|---|---|---|
| `syslog` | `rfc3164` | `udp` |
| `otlp` | `protobuf`, `json` | `grpc`, `http` |
| `cef`, `leef` | | `tcp`, `udp` |

### Date range replay

An Admin API mutation that replays a date range of audit log entries to a stream, to fill a gap or to load history when a stream is first configured.

A replayed entry carries the same `id` as the original, which the destination can use to detect a duplicate. Neither syslog nor a collector deduplicates.

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

### OpenTelemetry collectors

A collector is not a destination. It accepts syslog inside the customer network and forwards to one or more destinations. Bindplane is a distribution of it, and its Syslog source is the [syslog receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/syslogreceiver/README.md) of the OpenTelemetry Collector.

- It accepts any RFC 5424 message. The dialect and the framing have to match the sender.
- The structured data parameters become fields of the log record with no configuration.
- MSG stays a string. A `json_parser` is configured on it to turn the event object into fields.

### Observations

- The HTTPS APIs have no common format. They differ in both record grouping and envelope, so an HTTPS integration serves one service per implementation.
- Every service accepts syslog, usually through a collector that runs inside the customer network and forwards to the service over HTTPS. Google SecOps recommends the Bindplane agent, a distribution of the OpenTelemetry Collector, for this.
- CEF and LEEF are the only schemas parsed out of the box by more than one service. Both are flat `key=value`, so carrying the event object requires flattening it into dotted keys. Neither is supported at the moment. Either would be an additional `type`.

### How the design fits

- `type: syslog` reaches every service in the table through that service's collector, with no per-service implementation.
- An entry is JSON in MSG, the same object the portal shows as the Raw Event Log. The routable metadata is repeated in structured data, so a pipeline that does not parse MSG can still filter on `app_id`, `activity_type`, `user_id`, `client_id` and `ip_address`.
- Configuration is per project, as with Auth0 Log Streams and Okta Log Streaming.
- `type` names the encoding, so another encoding, such as OTLP, is added as a new `type` without changing existing configuration. See [additional types, transports and formats](#additional-types-transports-and-formats).
- Streaming carries only what is happening now. A gap is recoverable from the audit database, which retains every entry.
