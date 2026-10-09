# Portal Datadog Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let a project admin connect, edit and delete a Datadog audit log stream (site + API key) from the portal's Integrations page.

**Architecture:** The server gains a `set` action on the existing `telemetryAuditLogStreamSecrets` secret update instruction. The portal GraphQL API exposes it, plus which datadog streams have a key stored and the `telemetry` feature flag. The portal screen moves to `useAppSecretConfigForm`, so each Connect, Edit or Delete is one `updateApp` call that carries both the `authgear.yaml` change and the secret instruction. Stream-list manipulation lives in a pure, unit-tested helper module.

**Tech Stack:** Go (graphql-go, goconvey, YAML fixtures), React + TypeScript, Radix Themes, Apollo, jest.

**Spec:** [`2026-10-06-portal-datadog-design.md`](./2026-10-06-portal-datadog-design.md). Server behaviour: [`docs/specs/audit-log-streaming.md`](../../specs/audit-log-streaming.md).

## Global Constraints

- One app update per user action. Never send the stream and its key in separate calls: a `type: datadog` stream without its key makes the project fail to load.
- The API key is never returned by the portal API, masked or otherwise.
- Streams other than the managed one, and fields not in the dialog (`service`, `source`, `tags`), are preserved byte-for-byte in order.
- Managed stream = the first stream with `type: datadog`. New name: `datadog`, else `datadog-2`, `datadog-3`, ...
- Site options, in this order: `datadoghq.com` (US1, default), `us3.datadoghq.com` (US3), `us5.datadoghq.com` (US5), `datadoghq.eu` (EU1), `ap1.datadoghq.com` (AP1), `ap2.datadoghq.com` (AP2), `uk1.datadoghq.com` (UK1), `ddog-gov.com` (US1-FED), then Other.
- Datadog row hidden when `telemetry.audit_logs.streaming.disabled` is true; GTM row hidden when `google_tag_manager.disabled` is true; the route and nav item are shown when either row is visible.
- No breaking change. Nothing goes in `docs/BREAKING-CHANGES.md`.
- Do not hand-edit generated files: `schema.graphql`, `*.generated.ts`.
- Do not commit portal UI work mid-iteration. The final UI commit happens after the user has checked it visually (user preference).

## Review Focus

1. **Delete when the datadog stream is the only stream:** `cleanup` with `keepStreamNames: []` must remove the key and leave no empty `telemetry` object. Pinned in Task 1 (fixture) and Task 4 (helper test).
2. **A hand-written syslog stream already named `datadog`:** Connect must create `datadog-2` and leave the syslog stream untouched. Pinned in Task 4.
3. **Managed stream uses `http.endpoint` instead of `site`:** Edit must not write `datadog.site`, because the server rejects both together. Pinned in Task 4.
4. **Edit with the API key left blank:** no secret instruction is sent, so the stored key survives. Pinned in Task 4.
5. **Site stored that isn't in the list (e.g. `us2.ddog-gov.com`):** the dialog opens on Other with that value, and saving without touching it keeps it. Pinned in Task 4 (`siteToFormValue`).

---

## File map

| File | Responsibility |
|---|---|
| `pkg/lib/config/secret_update_instruction.go` | `set` action for `TelemetryAuditLogStreamSecretsUpdateInstruction`. |
| `pkg/lib/config/testdata/secret_update_instruction.yaml` | Fixtures for `set` and empty-list `cleanup`. |
| `pkg/portal/model/app.go` | `SecretConfig.TelemetryAuditLogStreamSecrets`; `PortalFeatureConfig.Telemetry`. |
| `pkg/portal/model/app_test.go` (new) | Read-side tests: stream names only, never `api_key`; feature flag passes through. |
| `pkg/portal/graphql/app.go` | GraphQL output types and the `SecretConfig` field. |
| `pkg/portal/graphql/app_mutation.go` | GraphQL input types and the `SecretConfigUpdateInstructionsInput` field. |
| `portal/src/types.ts` | TS types for app config `telemetry`, feature config `telemetry`, secrets, instruction. |
| `portal/src/graphql/portal/query/appAndSecretConfigQuery.graphql` | Select `telemetryAuditLogStreamSecrets { datadog { streamName } }`. |
| `portal/src/graphql/portal/integrations/datadog.ts` (new) | Pure helpers: site options, managed-stream lookup, name allocation, apply Connect/Edit/Delete. |
| `portal/src/graphql/portal/integrations/datadog.test.ts` (new) | jest tests for the helpers. |
| `portal/src/graphql/portal/IntegrationsConfigurationScreen.tsx` | Datadog row and dialog; switch to `useAppSecretConfigForm`. |
| `portal/src/AppRoot.tsx`, `portal/src/ScreenNav.tsx` | Gating. |
| `portal/src/locale-data/en.json` | Strings. |
| `portal/src/images/datadog_logo.svg` (new) | Row icon. |

---

### Task 1: `set` action for the telemetry secret instruction

**Files:**
- Modify: `pkg/lib/config/secret_update_instruction.go` (around lines 998-1051)
- Test: `pkg/lib/config/testdata/secret_update_instruction.yaml` (append)

**Interfaces:**
- Produces, JSON consumed by Task 2's GraphQL input:
  `{"telemetryAuditLogStreamSecrets": {"action": "set", "setData": {"datadog": [{"streamName": "...", "apiKey": "..."}]}}}`
  `{"telemetryAuditLogStreamSecrets": {"action": "cleanup", "cleanupData": {"keepStreamNames": []}}}`

- [ ] **Step 1: Write failing fixtures.** Append to `pkg/lib/config/testdata/secret_update_instruction.yaml`:

```yaml
---
name: set-telemetry-audit-log-stream-datadog-secret-add
error: null
currentSecretConfigYAML: |-
  secrets:
    - key: db
      data:
        database_url: "postgres://postgres@127.0.0.1:5432/postgres"
        database_schema: app
newSecretConfigYAML: |-
  secrets:
    - key: db
      data:
        database_url: "postgres://postgres@127.0.0.1:5432/postgres"
        database_schema: app
    - key: telemetry.audit_logs.streams.datadog
      data:
        - stream_name: "datadog"
          api_key: "new-key"
updateInstructionJSON: |-
  {
    "telemetryAuditLogStreamSecrets": {
      "action": "set",
      "setData": {
        "datadog": [{ "streamName": "datadog", "apiKey": "new-key" }]
      }
    }
  }
---
name: set-telemetry-audit-log-stream-datadog-secret-replace-keeps-others
error: null
currentSecretConfigYAML: |-
  secrets:
    - key: db
      data:
        database_url: "postgres://postgres@127.0.0.1:5432/postgres"
        database_schema: app
    - key: telemetry.audit_logs.streams.datadog
      data:
        - stream_name: "datadog"
          api_key: "old-key"
        - stream_name: "other"
          api_key: "other-key"
newSecretConfigYAML: |-
  secrets:
    - key: db
      data:
        database_url: "postgres://postgres@127.0.0.1:5432/postgres"
        database_schema: app
    - key: telemetry.audit_logs.streams.datadog
      data:
        - stream_name: "datadog"
          api_key: "new-key"
        - stream_name: "other"
          api_key: "other-key"
updateInstructionJSON: |-
  {
    "telemetryAuditLogStreamSecrets": {
      "action": "set",
      "setData": {
        "datadog": [{ "streamName": "datadog", "apiKey": "new-key" }]
      }
    }
  }
---
name: set-telemetry-audit-log-stream-datadog-secret-missing-set-data
error: |-
  config: missing setData for TelemetryAuditLogStreamSecretsUpdateInstruction
currentSecretConfigYAML: |-
  secrets:
    - key: db
      data:
        database_url: "postgres://postgres@127.0.0.1:5432/postgres"
        database_schema: app
newSecretConfigYAML: ""
updateInstructionJSON: |-
  {
    "telemetryAuditLogStreamSecrets": { "action": "set" }
  }
---
name: cleanup-telemetry-audit-log-stream-secrets-keep-none
error: null
currentSecretConfigYAML: |-
  secrets:
    - key: db
      data:
        database_url: "postgres://postgres@127.0.0.1:5432/postgres"
        database_schema: app
    - key: telemetry.audit_logs.streams.datadog
      data:
        - stream_name: "datadog"
          api_key: "old-key"
newSecretConfigYAML: |-
  secrets:
    - key: db
      data:
        database_url: "postgres://postgres@127.0.0.1:5432/postgres"
        database_schema: app
updateInstructionJSON: |-
  {
    "telemetryAuditLogStreamSecrets": {
      "action": "cleanup",
      "cleanupData": { "keepStreamNames": [] }
    }
  }
```

Also add a variant of the replace fixture where the current config has a `telemetry.audit_logs.streams.tls` item, and assert it comes out unchanged. Copy a tls block from the existing `cleanup-telemetry-audit-log-stream-secrets-keep-some` fixture.

Before relying on the error fixture, check how the harness compares `error`: read `pkg/lib/config/secret_update_instruction_test.go` after line 110. Match the format it expects.

- [ ] **Step 2: Run the tests and confirm they fail.**

Run: `go test ./pkg/lib/config/ -run TestSecretConfigUpdateInstruction`
Expected: FAIL on the `set-*` cases with `unexpected action for TelemetryAuditLogStreamSecretsUpdateInstruction: set`. The `keep-none` case may already pass, because a JSON `[]` decodes to a non-nil empty slice. Keep it as a regression pin.

- [ ] **Step 3: Implement.** In `secret_update_instruction.go`, next to the existing cleanup types:

```go
type TelemetryAuditLogStreamSecretsUpdateInstructionDatadogItem struct {
	StreamName string `json:"streamName,omitempty"`
	APIKey     string `json:"apiKey,omitempty"`
}

type TelemetryAuditLogStreamSecretsUpdateInstructionSetData struct {
	Datadog []TelemetryAuditLogStreamSecretsUpdateInstructionDatadogItem `json:"datadog,omitempty"`
}
```

Add `SetData *TelemetryAuditLogStreamSecretsUpdateInstructionSetData `json:"setData,omitempty"`` to `TelemetryAuditLogStreamSecretsUpdateInstruction`. Add `case SecretUpdateInstructionActionSet: return i.set(currentConfig)` to `ApplyTo`. Then:

```go
// set adds or replaces the datadog key of each named stream, leaving
// keys of other streams and the tls secret untouched.
func (i *TelemetryAuditLogStreamSecretsUpdateInstruction) set(currentConfig *SecretConfig) (*SecretConfig, error) {
	if i.SetData == nil {
		return nil, fmt.Errorf("config: missing setData for TelemetryAuditLogStreamSecretsUpdateInstruction")
	}

	out := &SecretConfig{}
	out.Secrets = make([]SecretItem, len(currentConfig.Secrets))
	copy(out.Secrets, currentConfig.Secrets)

	var items []TelemetryAuditLogStreamDatadogCredentialsItem
	idx, existing, found := out.Lookup(TelemetryAuditLogStreamDatadogCredentialsKey)
	if found {
		// RawData, not Data, for the reason given on pruneTelemetryAuditLogStreamSecretItems.
		if err := json.Unmarshal(existing.RawData, &items); err != nil {
			return nil, err
		}
	}

	for _, d := range i.SetData.Datadog {
		replaced := false
		for j := range items {
			if items[j].StreamName == d.StreamName {
				items[j].APIKey = d.APIKey
				replaced = true
			}
		}
		if !replaced {
			items = append(items, TelemetryAuditLogStreamDatadogCredentialsItem{
				StreamName: d.StreamName,
				APIKey:     d.APIKey,
			})
		}
	}

	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	newItem := SecretItem{Key: TelemetryAuditLogStreamDatadogCredentialsKey, RawData: json.RawMessage(data)}
	if found {
		out.Secrets[idx] = newItem
	} else {
		out.Secrets = append(out.Secrets, newItem)
	}
	return out, nil
}
```

Check the signature of `Lookup` (used by `pruneTelemetryAuditLogStreamSecretItems`): it returns `(int, *SecretItem, bool)`. Use it as that function does. Also update the instruction's doc comment, if it says only `cleanup` exists.

- [ ] **Step 4: Run the tests and confirm they pass.**

Run: `go test ./pkg/lib/config/...`
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add pkg/lib/config/secret_update_instruction.go pkg/lib/config/testdata/secret_update_instruction.yaml
git commit -m "[Telemetry] Add set action to the audit log stream secret update instruction"
```

---

### Task 2: Portal API: read stored datadog streams, write instruction, expose feature flag

**Files:**
- Modify: `pkg/portal/model/app.go` (`PortalFeatureConfig` ~line 14, `NewPortalFeatureConfig` ~line 38, `SecretConfig` ~line 158, `NewSecretConfig` ~line 181)
- Modify: `pkg/portal/graphql/app.go` (`AppSecretKey` consts ~line 246, `secretConfig` object ~line 258)
- Modify: `pkg/portal/graphql/app_mutation.go` (`secretConfigUpdateInstructionsInput` ~line 318)
- Create: `pkg/portal/model/app_test.go`
- Regenerate: `portal/src/graphql/portal/schema.graphql` via `make export-schemas`

**Interfaces:**
- Consumes: Task 1's instruction JSON shape.
- Produces GraphQL:
  - `SecretConfig.telemetryAuditLogStreamSecrets: TelemetryAuditLogStreamSecrets`
  - `TelemetryAuditLogStreamSecrets { datadog: [TelemetryAuditLogStreamDatadogSecret!] }`
  - `TelemetryAuditLogStreamDatadogSecret { streamName: String! }`
  - Input `SecretConfigUpdateInstructionsInput.telemetryAuditLogStreamSecrets: TelemetryAuditLogStreamSecretsUpdateInstructionsInput`, with `action: String!`, `setData { datadog: [{ streamName: String!, apiKey: String! }!]! }`, `cleanupData { keepStreamNames: [String!]! }`
  - `effectiveFeatureConfig.telemetry.audit_logs.streaming.disabled`

- [ ] **Step 1: Write a failing model test.** Create `pkg/portal/model/app_test.go`:

```go
package model

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

func TestNewSecretConfigTelemetry(t *testing.T) {
	Convey("NewSecretConfig telemetry audit log stream secrets", t, func() {
		ctx := context.Background()
		cfg, err := config.ParsePartialSecret(ctx, []byte(`
secrets:
- key: telemetry.audit_logs.streams.datadog
  data:
  - stream_name: datadog
    api_key: "super-secret-key"
`))
		So(err, ShouldBeNil)

		out, err := NewSecretConfig(cfg, []config.SecretKey{config.TelemetryAuditLogStreamDatadogCredentialsKey}, time.Now())
		So(err, ShouldBeNil)
		So(out.TelemetryAuditLogStreamSecrets, ShouldResemble, &TelemetryAuditLogStreamSecrets{
			Datadog: []TelemetryAuditLogStreamDatadogSecret{{StreamName: "datadog"}},
		})

		b, err := json.Marshal(out)
		So(err, ShouldBeNil)
		So(string(b), ShouldNotContainSubstring, "super-secret-key")
	})

	Convey("NewPortalFeatureConfig passes telemetry through", t, func() {
		disabled := true
		fc := &config.FeatureConfig{Telemetry: &config.TelemetryFeatureConfig{
			AuditLogs: &config.TelemetryAuditLogsFeatureConfig{
				Streaming: &config.TelemetryAuditLogsStreamingFeatureConfig{Disabled: &disabled},
			},
		}}
		So(*NewPortalFeatureConfig(fc).Telemetry.AuditLogs.Streaming.Disabled, ShouldBeTrue)
	})
}
```

Check the actual import path of goconvey and the module path in another `_test.go` under `pkg/portal`, and match them.

- [ ] **Step 2: Run the test and confirm it fails.**

Run: `go test ./pkg/portal/model/ -run TestNewSecretConfigTelemetry`
Expected: compile failure, `out.TelemetryAuditLogStreamSecrets undefined`.

- [ ] **Step 3: Implement the model.** In `pkg/portal/model/app.go`:

```go
type TelemetryAuditLogStreamDatadogSecret struct {
	StreamName string `json:"streamName,omitempty"`
}

type TelemetryAuditLogStreamSecrets struct {
	Datadog []TelemetryAuditLogStreamDatadogSecret `json:"datadog,omitempty"`
}
```

Add `TelemetryAuditLogStreamSecrets *TelemetryAuditLogStreamSecrets `json:"telemetryAuditLogStreamSecrets,omitempty"`` to `SecretConfig`. Add `Telemetry *config.TelemetryFeatureConfig `json:"telemetry,omitempty"`` to `PortalFeatureConfig`, and `Telemetry: c.Telemetry,` in `NewPortalFeatureConfig`. In `NewSecretConfig`, before `return out, nil`:

```go
	// The API key is write-only through the portal; only the stream names are exposed.
	if creds, ok := secretConfig.LookupData(config.TelemetryAuditLogStreamDatadogCredentialsKey).(*config.TelemetryAuditLogStreamDatadogCredentials); ok {
		secrets := &TelemetryAuditLogStreamSecrets{}
		for _, item := range *creds {
			secrets.Datadog = append(secrets.Datadog, TelemetryAuditLogStreamDatadogSecret{StreamName: item.StreamName})
		}
		out.TelemetryAuditLogStreamSecrets = secrets
	}
```

- [ ] **Step 4: Implement the GraphQL types.** In `pkg/portal/graphql/app.go`, add next to `botProtectionProviderSecret`:

```go
var telemetryAuditLogStreamDatadogSecret = graphql.NewObject(graphql.ObjectConfig{
	Name: "TelemetryAuditLogStreamDatadogSecret",
	Fields: graphql.Fields{
		"streamName": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
	},
})

var telemetryAuditLogStreamSecrets = graphql.NewObject(graphql.ObjectConfig{
	Name: "TelemetryAuditLogStreamSecrets",
	Fields: graphql.Fields{
		"datadog": &graphql.Field{
			Type: graphql.NewList(graphql.NewNonNull(telemetryAuditLogStreamDatadogSecret)),
		},
	},
})
```

Add `AppSecretKeyTelemetryAuditLogStreamSecrets AppSecretKey = "telemetryAuditLogStreamSecrets" // nolint:gosec` and `string(AppSecretKeyTelemetryAuditLogStreamSecrets): &graphql.Field{Type: telemetryAuditLogStreamSecrets},` in `secretConfig`. Do **not** add it to the `appSecretKey` enum: nothing here can be revealed, so a visit token for it would have no effect.

In `pkg/portal/graphql/app_mutation.go`, add before `secretConfigUpdateInstructionsInput`:

```go
var telemetryAuditLogStreamDatadogSecretInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "TelemetryAuditLogStreamDatadogSecretInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"streamName": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"apiKey":     &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
	},
})

var telemetryAuditLogStreamSecretsSetDataInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "TelemetryAuditLogStreamSecretsSetDataInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"datadog": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(telemetryAuditLogStreamDatadogSecretInput))),
		},
	},
})

var telemetryAuditLogStreamSecretsCleanupDataInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "TelemetryAuditLogStreamSecretsCleanupDataInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"keepStreamNames": &graphql.InputObjectFieldConfig{
			Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.String))),
		},
	},
})

var telemetryAuditLogStreamSecretsUpdateInstructionsInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "TelemetryAuditLogStreamSecretsUpdateInstructionsInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"action":      &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
		"setData":     &graphql.InputObjectFieldConfig{Type: telemetryAuditLogStreamSecretsSetDataInput},
		"cleanupData": &graphql.InputObjectFieldConfig{Type: telemetryAuditLogStreamSecretsCleanupDataInput},
	},
})
```

Then add `"telemetryAuditLogStreamSecrets": &graphql.InputObjectFieldConfig{Type: telemetryAuditLogStreamSecretsUpdateInstructionsInput},` to `secretConfigUpdateInstructionsInput`.

- [ ] **Step 5: Run the tests and regenerate the schema.**

Run: `go test ./pkg/portal/... && make export-schemas`
Expected: PASS. `git diff portal/src/graphql/portal/schema.graphql` shows the new types. Use the `generate-schemas-and-gentype` skill if `make export-schemas` fails.

- [ ] **Step 6: Commit.**

```bash
git add pkg/portal portal/src/graphql/portal/schema.graphql
git commit -m "[Portal] Expose audit log stream datadog secrets and telemetry feature config"
```

---

### Task 3: Portal types, query, codegen and gating

**Files:**
- Modify: `portal/src/types.ts` (`PortalAPIAppConfig` ~797, `PortalAPISecretConfig` ~893, `PortalAPISecretConfigUpdateInstruction` ~985, `PortalAPIFeatureConfig` ~1003)
- Modify: `portal/src/graphql/portal/query/appAndSecretConfigQuery.graphql`
- Modify: `portal/src/AppRoot.tsx:31-32`, `portal/src/ScreenNav.tsx:146-148`
- Regenerate: `*.generated.ts` via `npm run gentype`

**Interfaces:**
- Produces TS types used by Task 4 and Task 5: `TelemetryConfig`, `TelemetryAuditLogStreamConfig`, `TelemetryAuditLogStreamSecretsUpdateInstruction`, `PortalAPISecretConfig.telemetryAuditLogStreamSecrets`, `PortalAPIFeatureConfig.telemetry`, and an exported `isIntegrationsAvailable(fc)` in `AppRoot.tsx` reused by `ScreenNav.tsx`.

- [ ] **Step 1: Add types** to `portal/src/types.ts`:

```ts
export interface TelemetryAuditLogStreamDatadogConfig {
  site?: string;
  service?: string;
  source?: string;
  tags?: Record<string, string>;
}

export interface TelemetryAuditLogStreamHTTPConfig {
  endpoint?: string;
}

// Only the fields the portal reads or writes are typed; anything else
// on a stream is kept by spreading the original object.
export interface TelemetryAuditLogStreamConfig {
  name: string;
  type: "syslog" | "datadog";
  transport: "tcp" | "http";
  datadog?: TelemetryAuditLogStreamDatadogConfig;
  http?: TelemetryAuditLogStreamHTTPConfig;
  [key: string]: unknown;
}

export interface TelemetryAuditLogsConfig {
  streams?: TelemetryAuditLogStreamConfig[];
}

export interface TelemetryConfig {
  audit_logs?: TelemetryAuditLogsConfig;
}

export interface TelemetryFeatureConfig {
  audit_logs?: { streaming?: { disabled?: boolean } };
}

export interface TelemetryAuditLogStreamSecrets {
  datadog?: { streamName: string }[] | null;
}

export interface TelemetryAuditLogStreamSecretsUpdateInstruction {
  action: "set" | "cleanup";
  setData?: { datadog: { streamName: string; apiKey: string }[] } | null;
  cleanupData?: { keepStreamNames: string[] } | null;
}
```

Add `telemetry?: TelemetryConfig;` to `PortalAPIAppConfig`, `telemetry?: TelemetryFeatureConfig;` to `PortalAPIFeatureConfig`, `telemetryAuditLogStreamSecrets?: TelemetryAuditLogStreamSecrets | null;` to `PortalAPISecretConfig`, and `telemetryAuditLogStreamSecrets?: TelemetryAuditLogStreamSecretsUpdateInstruction | null;` to `PortalAPISecretConfigUpdateInstruction`.

- [ ] **Step 2: Query.** In `appAndSecretConfigQuery.graphql`, inside `secretConfig(token: $token) { ... }`, after `smsProviderSecrets { ... }`:

```graphql
    telemetryAuditLogStreamSecrets {
      datadog {
        streamName
      }
    }
```

- [ ] **Step 3: Gating.** In `AppRoot.tsx`, replace `isIntegrationsAvailable` and export it:

```ts
export const isGoogleTagManagerAvailable = (
  fc: PortalAPIFeatureConfig | null
): boolean => (fc?.google_tag_manager?.disabled ?? false) === false;
export const isAuditLogStreamingAvailable = (
  fc: PortalAPIFeatureConfig | null
): boolean => (fc?.telemetry?.audit_logs?.streaming?.disabled ?? false) === false;
const isIntegrationsAvailable = (fc: PortalAPIFeatureConfig | null): boolean =>
  isGoogleTagManagerAvailable(fc) || isAuditLogStreamingAvailable(fc);
```

If `ScreenNav.tsx` importing from `AppRoot.tsx` creates an import cycle (AppRoot imports ScreenNav indirectly), move these three functions into a new `portal/src/util/integrations.ts` instead, and import them from both files. In `ScreenNav.tsx:146`, set `showIntegrations` from the same predicate, applied to `app?.effectiveFeatureConfig`. Check the generated type of `effectiveFeatureConfig` there: it is a GraphQL scalar, so a cast to `PortalAPIFeatureConfig` may be needed, as elsewhere in the file.

- [ ] **Step 4: Codegen and typecheck.**

Run: `cd portal && npm run gentype && npm run typecheck`
Expected: no errors. `appAndSecretConfigQuery.generated.ts` includes `telemetryAuditLogStreamSecrets`.

- [ ] **Step 5: Commit.**

```bash
git add portal/src/types.ts portal/src/graphql/portal/query portal/src/AppRoot.tsx portal/src/ScreenNav.tsx portal/src/util
git commit -m "[Portal] Add telemetry types and gate Integrations on audit log streaming"
```

---

### Task 4: Pure Datadog stream helpers

**Files:**
- Create: `portal/src/graphql/portal/integrations/datadog.ts`
- Test: `portal/src/graphql/portal/integrations/datadog.test.ts`

**Interfaces:**
- Consumes: Task 3 types.
- Produces, for Task 5:

```ts
export const DATADOG_DEFAULT_SITE = "datadoghq.com";
export const DATADOG_SITE_OTHER = "__other__";
export interface DatadogSiteOption { site: string; label: string } // label e.g. "US1"
export const DATADOG_SITE_OPTIONS: DatadogSiteOption[];
export function findManagedDatadogStream(config: PortalAPIAppConfig): TelemetryAuditLogStreamConfig | null;
export function allocateDatadogStreamName(config: PortalAPIAppConfig): string;
export function siteToFormValue(site: string | undefined): { option: string; other: string };
export function formValueToSite(option: string, other: string): string;
export function applyDatadogConnect(config: PortalAPIAppConfig, site: string): PortalAPIAppConfig; // returns new config; stream name via allocateDatadogStreamName
export function applyDatadogEdit(config: PortalAPIAppConfig, site: string): PortalAPIAppConfig;   // no-op on site when managed stream has http.endpoint
export function applyDatadogDelete(config: PortalAPIAppConfig): PortalAPIAppConfig;                // drops empty streams/audit_logs/telemetry
export function remainingStreamNames(config: PortalAPIAppConfig): string[];
```

- [ ] **Step 1: Write failing tests** in `datadog.test.ts`:

```ts
import { describe, it, expect } from "@jest/globals";
import { PortalAPIAppConfig } from "../../../types";
import {
  allocateDatadogStreamName,
  applyDatadogConnect,
  applyDatadogDelete,
  applyDatadogEdit,
  findManagedDatadogStream,
  formValueToSite,
  remainingStreamNames,
  siteToFormValue,
  DATADOG_SITE_OTHER,
} from "./datadog";

const syslog = {
  name: "datadog",
  type: "syslog" as const,
  transport: "tcp" as const,
  tcp: { address: "collector:5140" },
  syslog: { format: "rfc5424", framing: "newline" },
};

describe("datadog helpers", () => {
  it("connect appends a stream and keeps others", () => {
    const cfg: PortalAPIAppConfig = { id: "app", telemetry: { audit_logs: { streams: [syslog] } } };
    const out = applyDatadogConnect(cfg, "datadoghq.eu");
    expect(out.telemetry?.audit_logs?.streams).toEqual([
      syslog,
      { name: "datadog-2", type: "datadog", transport: "http", datadog: { site: "datadoghq.eu" } },
    ]);
    expect(cfg.telemetry?.audit_logs?.streams).toHaveLength(1); // input not mutated
  });

  it("allocates datadog when free", () => {
    expect(allocateDatadogStreamName({ id: "app" })).toBe("datadog");
  });

  it("edit preserves service/source/tags", () => {
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: { audit_logs: { streams: [
        { name: "dd", type: "datadog", transport: "http", datadog: { site: "datadoghq.com", service: "s", tags: { env: "prod" } } },
      ] } },
    };
    const out = applyDatadogEdit(cfg, "us5.datadoghq.com");
    expect(findManagedDatadogStream(out)?.datadog).toEqual({ site: "us5.datadoghq.com", service: "s", tags: { env: "prod" } });
  });

  it("edit does not set site when http.endpoint is used", () => {
    const stream = { name: "dd", type: "datadog" as const, transport: "http" as const, http: { endpoint: "https://opw.internal/api/v2/logs" } };
    const cfg: PortalAPIAppConfig = { id: "app", telemetry: { audit_logs: { streams: [stream] } } };
    expect(findManagedDatadogStream(applyDatadogEdit(cfg, "datadoghq.eu"))).toEqual(stream);
  });

  it("delete of the only stream removes telemetry entirely", () => {
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: { audit_logs: { streams: [{ name: "datadog", type: "datadog", transport: "http" }] } },
    };
    const out = applyDatadogDelete(cfg);
    expect(out.telemetry).toBeUndefined();
    expect(remainingStreamNames(out)).toEqual([]);
  });

  it("delete keeps other streams", () => {
    const cfg: PortalAPIAppConfig = {
      id: "app",
      telemetry: { audit_logs: { streams: [syslog, { name: "datadog-2", type: "datadog", transport: "http" }] } },
    };
    expect(remainingStreamNames(applyDatadogDelete(cfg))).toEqual(["datadog"]);
  });

  it("unknown site round-trips through Other", () => {
    expect(siteToFormValue("us2.ddog-gov.com")).toEqual({ option: DATADOG_SITE_OTHER, other: "us2.ddog-gov.com" });
    expect(siteToFormValue(undefined)).toEqual({ option: "datadoghq.com", other: "" });
    expect(formValueToSite(DATADOG_SITE_OTHER, "  us2.ddog-gov.com ")).toBe("us2.ddog-gov.com");
    expect(formValueToSite("datadoghq.eu", "ignored")).toBe("datadoghq.eu");
  });
});
```

- [ ] **Step 2: Run the tests and confirm they fail.**

Run: `cd portal && npx jest src/graphql/portal/integrations/datadog.test.ts`
Expected: FAIL, `Cannot find module './datadog'`.

- [ ] **Step 3: Implement** `datadog.ts`:

```ts
import { produce } from "immer";
import {
  PortalAPIAppConfig,
  TelemetryAuditLogStreamConfig,
} from "../../../types";
import { clearEmptyObject } from "../../../util/misc";

export const DATADOG_DEFAULT_SITE = "datadoghq.com";
export const DATADOG_SITE_OTHER = "__other__";
const DATADOG_STREAM_NAME = "datadog";

export interface DatadogSiteOption {
  site: string;
  label: string;
}

export const DATADOG_SITE_OPTIONS: DatadogSiteOption[] = [
  { site: "datadoghq.com", label: "US1" },
  { site: "us3.datadoghq.com", label: "US3" },
  { site: "us5.datadoghq.com", label: "US5" },
  { site: "datadoghq.eu", label: "EU1" },
  { site: "ap1.datadoghq.com", label: "AP1" },
  { site: "ap2.datadoghq.com", label: "AP2" },
  { site: "uk1.datadoghq.com", label: "UK1" },
  { site: "ddog-gov.com", label: "US1-FED" },
];

function streamsOf(config: PortalAPIAppConfig): TelemetryAuditLogStreamConfig[] {
  return config.telemetry?.audit_logs?.streams ?? [];
}

export function findManagedDatadogStream(
  config: PortalAPIAppConfig
): TelemetryAuditLogStreamConfig | null {
  return streamsOf(config).find((s) => s.type === "datadog") ?? null;
}

export function allocateDatadogStreamName(config: PortalAPIAppConfig): string {
  const taken = new Set(streamsOf(config).map((s) => s.name));
  if (!taken.has(DATADOG_STREAM_NAME)) {
    return DATADOG_STREAM_NAME;
  }
  for (let n = 2; ; n++) {
    const name = `${DATADOG_STREAM_NAME}-${n}`;
    if (!taken.has(name)) {
      return name;
    }
  }
}

export function siteToFormValue(site: string | undefined): {
  option: string;
  other: string;
} {
  const value = site ?? DATADOG_DEFAULT_SITE;
  return DATADOG_SITE_OPTIONS.some((o) => o.site === value)
    ? { option: value, other: "" }
    : { option: DATADOG_SITE_OTHER, other: value };
}

export function formValueToSite(option: string, other: string): string {
  return option === DATADOG_SITE_OTHER ? other.trim() : option;
}

export function applyDatadogConnect(
  config: PortalAPIAppConfig,
  site: string
): PortalAPIAppConfig {
  const name = allocateDatadogStreamName(config);
  return produce(config, (draft) => {
    draft.telemetry ??= {};
    draft.telemetry.audit_logs ??= {};
    draft.telemetry.audit_logs.streams ??= [];
    draft.telemetry.audit_logs.streams.push({
      name,
      type: "datadog",
      transport: "http",
      datadog: { site },
    });
  });
}

export function applyDatadogEdit(
  config: PortalAPIAppConfig,
  site: string
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    const stream = draft.telemetry?.audit_logs?.streams?.find(
      (s) => s.type === "datadog"
    );
    // site and http.endpoint are mutually exclusive on the server.
    if (stream == null || stream.http?.endpoint != null) {
      return;
    }
    stream.datadog ??= {};
    stream.datadog.site = site;
  });
}

export function applyDatadogDelete(
  config: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    const streams = draft.telemetry?.audit_logs?.streams;
    if (streams == null) {
      return;
    }
    const idx = streams.findIndex((s) => s.type === "datadog");
    if (idx >= 0) {
      streams.splice(idx, 1);
    }
    if (streams.length === 0) {
      delete draft.telemetry!.audit_logs!.streams;
    }
    clearEmptyObject(draft);
  });
}

export function remainingStreamNames(config: PortalAPIAppConfig): string[] {
  return streamsOf(config).map((s) => s.name);
}
```

Before relying on `clearEmptyObject`, check `portal/src/util/misc.ts` to confirm that it removes nested empty objects (`telemetry: { audit_logs: {} }`). If it only clears one level, delete `audit_logs` and `telemetry` explicitly when they are empty. Avoid the `!` non-null assertions if lint forbids them, by narrowing with local variables instead.

- [ ] **Step 4: Run the tests and confirm they pass.**

Run: `cd portal && npx jest src/graphql/portal/integrations/datadog.test.ts && npm run typecheck`
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add portal/src/graphql/portal/integrations
git commit -m "[Portal] Add Datadog audit log stream helpers"
```

---

### Task 5: Integrations screen: Datadog row and dialog

**Files:**
- Modify: `portal/src/graphql/portal/IntegrationsConfigurationScreen.tsx`
- Modify: `portal/src/graphql/portal/IntegrationsConfigurationScreen.module.css` (only if the Other field needs spacing)
- Modify: `portal/src/locale-data/en.json` (next to the `IntegrationsConfigurationScreen.*` keys, ~line 2212)
- Create: `portal/src/images/datadog_logo.svg`

Load the `update-portal-ui` skill before this task, and follow its link, i18n and hardcoded-text rules.

**Interfaces:**
- Consumes: Task 4 helpers; `useAppSecretConfigForm` (`hook/useAppSecretConfigForm.ts`), with `secretVisitToken: null`; `useAppFeatureConfigQuery(appID)` for the row gates; `isGoogleTagManagerAvailable` / `isAuditLogStreamingAvailable` from Task 3.

**Form state:**

```ts
interface FormState {
  googleTagManagerContainerID: string;
  datadog: {
    connected: boolean;       // managed stream exists AND has a stored key
    usesEndpoint: boolean;    // managed stream has http.endpoint
    endpoint: string;
    siteOption: string;       // from siteToFormValue
    siteOther: string;
    apiKey: string;           // always "" from constructFormState; never prefilled
    action: "none" | "connect" | "edit" | "delete";
  };
}
```

- `constructFormState(config, secrets)`:
  - `connected` = `findManagedDatadogStream(config) != null` **and** that stream's name appears in `secrets.telemetryAuditLogStreamSecrets?.datadog`. If the stream has no stored key, the project could not load at all, so in practice the two agree.
  - `config` here is the *effective* config, which has defaults applied. Read site and endpoint from it.
- `constructConfig(raw, secrets, initial, current)`:
  - Always start from `raw` (not effective), apply the GTM change as today, then branch on `current.datadog.action`:
    - `connect`: `applyDatadogConnect(raw, site)`
    - `edit`: `applyDatadogEdit(raw, site)`
    - `delete`: `applyDatadogDelete(raw)`
    - `none`: unchanged
  - Return `[newConfig, secrets]`.
- `constructSecretUpdateInstruction(newConfig, _secrets, current)`:
  - `connect`: `{ telemetryAuditLogStreamSecrets: { action: "set", setData: { datadog: [{ streamName: findManagedDatadogStream(newConfig)!.name, apiKey: current.datadog.apiKey.trim() }] } } }`
  - `edit` with a non-blank `apiKey`: the same `set`, using the managed stream's name.
  - `edit` with a blank `apiKey`: `undefined`.
  - `delete`: `{ telemetryAuditLogStreamSecrets: { action: "cleanup", cleanupData: { keepStreamNames: remainingStreamNames(newConfig) } } }`
  - `none`: `undefined`

The GTM handlers keep their behaviour. `form.saveWith(fn)` from `useAppConfigForm` becomes `form.saveWithState({ ...form.state, ... })` on the new hook. Check the exact signature in `useAppSecretConfigForm.ts` (`saveWithState(state, ignoreConflict?)`), and how `state` is initialised before it is set.

- [ ] **Step 1: Strings.** Add to `en.json`:

```json
"IntegrationsConfigurationScreen.add-on.datadog.name": "Datadog",
"IntegrationsConfigurationScreen.add-on.datadog.description": "Stream audit logs to your Datadog organization.",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.title": "Connect Datadog",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.site.label": "Datadog site",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.site.option": "{label} ({site})",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.site.other": "Other",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.site.other.placeholder": "e.g. us2.ddog-gov.com",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.endpoint.label": "Endpoint",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.api-key.label": "API key",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.api-key.hint": "Use an API key, not an application key. Find it in Datadog under <ExternalLink>Organization Settings > API Keys</ExternalLink>.",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.api-key.keep-hint": "Leave blank to keep the current key.",
"IntegrationsConfigurationScreen.add-on.datadog.dialog.delivery-note": "Logs appear in Datadog within about a minute. An incorrect key or site is not detected when saving."
```

The API keys link target is `https://app.datadoghq.com/organization-settings/api-keys`. Render it with the inline-link pattern from the `update-portal-ui` skill, not a hardcoded `<a>` in the string.

- [ ] **Step 2: Logo.** Add the Datadog logo as `portal/src/images/datadog_logo.svg`, from Datadog's press kit (the purple dog mark). Import it as `gtmLogoURL` is imported.

- [ ] **Step 3: Implement the screen.**
  - Switch `IntegrationsConfigurationScreen` to `useAppSecretConfigForm({ appID, secretVisitToken: null, constructFormState, constructConfig, constructSecretUpdateInstruction })`.
  - Build `items` with an `id: "gtm" | "datadog"`. Filter by the feature gates, using `useAppFeatureConfigQuery(appID).effectiveFeatureConfig`.
  - Each row's action button opens its own dialog. Today every row opens the GTM dialog, so switch the handler on `item.id`.
  - Datadog dialog, structured like the GTM dialog:
    - **Site:** a Radix `Select.Root` / `Select.Trigger` / `Select.Content` / `Select.Item` (see `EditCustomAttributeForm.tsx:308` for usage). Its items are `DATADOG_SITE_OPTIONS`, labelled `{label} ({site})`, plus Other. When Other is selected, a `TextField` for `siteOther` follows. When `usesEndpoint` is true, show a read-only `TextField` with the endpoint instead of the select.
    - **API key:** `TextField type="password"`. Its `hint` is the API-key hint, plus the keep-hint when `connected`.
    - **Delivery note:** `Text size="1"` below the fields.
    - **Buttons:** as in the GTM dialog. Delete is shown only when `connected`.
  - Client-side validation on Save:
    - Not connected and `apiKey.trim() === ""` → `errors.validation.required` on the API key.
    - `siteOption === DATADOG_SITE_OTHER && siteOther.trim() === ""` → `errors.validation.required` on the other-site field.
  - Server errors: match `parentJSONPointer: /^\/telemetry\/audit_logs\/streams\/\d+\/datadog$/`, `fieldName: "site"` onto the site field, the same way `containerIDError` is computed today. Show any other `updateError` with `ErrorRenderer` at the top of the dialog.
  - Saving: `form.saveWithState({ ...state, datadog: { ...draft, action } })`, where `action` is `connect` when not connected, otherwise `edit`. Delete uses `action: "delete"`. Close the dialog on success.

- [ ] **Step 4: Static checks.**

Run: `cd portal && npm run typecheck && make -C . lint && npx jest src/graphql/portal/integrations`
Expected: all pass. Fix lint findings. Do not disable rules.

- [ ] **Step 5: Hand off for visual check.** Do not commit yet. Tell the user the UI is ready to check at `/project/<appID>/integrations`, with the local stack running (see `CLAUDE.md`, "Start local dev"). Ask them to check:
  1. Connect with US1 creates a stream and the row shows Connected.
  2. Edit to EU1 with a blank key keeps the key: `authgear.secrets.yaml` is unchanged in `git diff` / config source.
  3. Delete removes both.
  4. With `telemetry.audit_logs.streaming.disabled: true` (see the "Edit feature config for a project" memory for how to flip it), the row is hidden.

- [ ] **Step 6: Commit after the user approves.**

```bash
git add portal/src/graphql/portal/IntegrationsConfigurationScreen.tsx portal/src/graphql/portal/IntegrationsConfigurationScreen.module.css portal/src/locale-data/en.json portal/src/images/datadog_logo.svg
git commit -m "[Portal] Connect Datadog audit log streaming from Integrations"
```

---

### Task 6: Verification gate

- [ ] **Step 1:** Run `go test ./pkg/lib/config/... ./pkg/portal/...`, `make lint`, and `cd portal && npm run typecheck && npm test`. Expected: all pass.
- [ ] **Step 2:** If any goanalysis line positions moved, run `make update-vettedpositions` (skill `update-vettedpositions`) and commit as `chore: Update .vettedpositions`.
- [ ] **Step 3:** Run the `review-pr` skill on the branch. Resolve every Bugs, Security, Performance and Code quality finding, or state why it is a false positive.
- [ ] **Step 4:** Push to the `fork` remote and open a PR against `authgear/authgear-server:main`, titled `[Portal] Connect Datadog audit log streaming from Integrations`. Only do this when the user asks.
