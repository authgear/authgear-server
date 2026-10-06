# Audit Log Streaming, type datadog — Portal Integrations design

Lets a project admin connect the project's audit log stream to their own Datadog organization from the portal's Integrations page. The server side is [`docs/specs/audit-log-streaming.md`](../../specs/audit-log-streaming.md); this design adds only the portal surface on top of it.

## Goal / scope

- The Integrations page lists **Datadog** next to Google Tag Manager, with Connect, Edit and Delete.
- The admin sets the Datadog **site** and **API key**. Nothing else.
- Out of scope, YAML only: `service`, `source`, `tags`, `http.endpoint`, more than one datadog stream, and syslog streams.
- No breaking change. Every API change is additive.

## The managed stream

The portal manages exactly one stream: the first item of `telemetry.audit_logs.streams` with `type: datadog`.

| Situation | Behaviour |
|---|---|
| No datadog stream | The row shows Connect. Connect appends `{name, type: datadog, transport: http, datadog: {site}}`. |
| New stream name | `datadog`, or `datadog-2`, `datadog-3`, ... when taken by another stream. |
| Managed stream exists | The row shows Connected and Edit. Edit changes `datadog.site` only. |
| Every other stream | Passed through unchanged, in its original position. |
| Fields not in the dialog (`service`, `source`, `tags`) | Preserved on Edit. |
| Managed stream has `http.endpoint` | Edit shows the endpoint read-only and hides the site picker, because `site` and `endpoint` are mutually exclusive. Delete still works. |
| Site equal to the default `datadoghq.com` | Written explicitly, so the YAML states the choice. |

## Feature gating

| Flag | Effect |
|---|---|
| `google_tag_manager.disabled` | Hides the GTM row. |
| `telemetry.audit_logs.streaming.disabled` | Hides the Datadog row. |

The Integrations route and nav item are available when at least one row is visible.

## Portal API changes

### Read: `SecretConfig`

```graphql
type SecretConfig {
  # existing fields ...
  telemetryAuditLogStreamSecrets: TelemetryAuditLogStreamSecrets
}

type TelemetryAuditLogStreamSecrets {
  datadog: [TelemetryAuditLogStreamDatadogSecret!]
}

type TelemetryAuditLogStreamDatadogSecret {
  streamName: String!
  apiKey: String
}

enum AppSecretKey {
  # existing values ...
  TELEMETRY_AUDIT_LOG_STREAM_SECRETS
}
```

`apiKey` is `null` unless the query carries a secret visit token for `TELEMETRY_AUDIT_LOG_STREAM_SECRETS`, which requires reauthentication, the same as `BOT_PROTECTION_PROVIDER_SECRET`. The token reveals the datadog API keys only, never the tls key.

### Write: `SecretConfigUpdateInstructionsInput`

```graphql
input SecretConfigUpdateInstructionsInput {
  # existing fields ...
  telemetryAuditLogStreamSecrets: TelemetryAuditLogStreamSecretsUpdateInstructionsInput
}

input TelemetryAuditLogStreamSecretsUpdateInstructionsInput {
  action: String!            # "set" | "cleanup"
  setData: TelemetryAuditLogStreamSecretsSetDataInput
  cleanupData: TelemetryAuditLogStreamSecretsCleanupDataInput
}

input TelemetryAuditLogStreamSecretsSetDataInput {
  datadog: [TelemetryAuditLogStreamDatadogSecretInput!]!
}

input TelemetryAuditLogStreamDatadogSecretInput {
  streamName: String!
  apiKey: String!
}

input TelemetryAuditLogStreamSecretsCleanupDataInput {
  keepStreamNames: [String!]!
}
```

### The `set` action

`TelemetryAuditLogStreamSecretsUpdateInstruction` gains `action: set`.

- For each item in `setData.datadog`, the item of `telemetry.audit_logs.streams.datadog` with the same `streamName` is replaced. If none exists, the item is appended.
- Items of other streams, and `telemetry.audit_logs.streams.tls`, are untouched.
- `cleanup` keeps its current meaning.
- `keepStreamNames: []` is valid and means "remove every item". An empty list must survive the trip from GraphQL to the instruction. Today `KeepStreamNames` is `omitempty`, so an empty list can be dropped and rejected as missing.

## Save flows

Every flow is **one** app update that carries both the `authgear.yaml` change and the secret instruction. The server validates a datadog stream and its key as a pair, so splitting them across two saves would leave the project unloadable in between.

| Action | `authgear.yaml` | Secret instruction |
|---|---|---|
| Connect | Append the managed stream. | `set` with `{streamName, apiKey}`. |
| Edit, API key unchanged | Update `datadog.site`. | None. |
| Edit, API key changed | Update `datadog.site`. | `set` with the new key. |
| Delete | Remove the managed stream. | `cleanup` with the names of every remaining stream. |

## UI

**Row:** Datadog logo, name "Datadog", and the description "Stream audit logs to your Datadog organization." Once connected, it shows the Connected badge and the action changes to Edit, the same as the GTM row.

**Dialog**, the same structure as the GTM dialog:

- **Site:** a dropdown of US1 `datadoghq.com` (the default), US3 `us3.datadoghq.com`, US5 `us5.datadoghq.com`, EU1 `datadoghq.eu`, AP1 `ap1.datadoghq.com`, AP2 `ap2.datadoghq.com`, UK1 `uk1.datadoghq.com`, US1-FED `ddog-gov.com`, and **Other**. Choosing Other reveals a text field for any site parameter. A stored site not in the list opens as Other with that value filled in.
- **API key:** a text field.
  - **On Connect:** empty and required.
  - **On Edit:** read-only, with an Edit button. After reauthentication it shows the current key; before, a masked value. Edit on a masked key reauthenticates first, then reopens the dialog with the unsaved site choice kept and the key revealed. Once unlocked, the key is required.
  - **Help text:** this must be an API key, not an application key, with a link to Datadog's API keys page.
- **Note:** "Logs appear in Datadog within about a minute." A wrong key or site is only detected when logs are delivered, never when saving.
- **Buttons:** Delete (only when connected), Cancel, Save.

**Validation:**

- **In the browser:** API key non-blank on Connect and on Edit; Other site non-empty after trimming.
- **From the server:** errors on `/telemetry/audit_logs/streams/<i>/datadog/site` show on the site field. Any other error shows at the top of the dialog.

The screen moves from `useAppConfigForm` to `useAppSecretConfigForm`. GTM behaviour does not change.

## Testing

- **Config tests:** cases for the `set` action in `pkg/lib/config/testdata/secret_update_instruction.yaml`: add, replace, other streams untouched, tls untouched, and `cleanup` with `keepStreamNames: []`.
- **Portal model tests:** `SecretConfig` lists datadog stream names, and includes the API key only when unmasked.
- **Generated files:** regenerated with `make export-schemas` and portal `npm run gentype`.
- **Portal checks:** `npm run typecheck` and `make -C portal lint`.
- **e2e:** no new test. Delivery is already covered by `e2e/tests/audit_log_streaming_datadog`.
