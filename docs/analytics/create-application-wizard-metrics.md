# Create Application Wizard — PostHog Report

Companion to `docs/superpowers/specs/2026-07-09-create-application-wizard-metrics-design.md`.
All data is in PostHog; no manual SQL against the audit DB is needed.

## Events

| Event | Source | Key properties |
|---|---|---|
| `createApplication.viewed` | portal (GTM) | `wizard_version` |
| `createApplication.selected-type` | portal (GTM) | `application_type`, `wizard_version` |
| `createApplication.created` | portal (GTM) | `client_id`, `application_type`, `wizard_version` |
| `application.first_auth` | auth server (real-time event sink) | `client_id`, `app_id` |

## Insight A — Wizard completion / drop-off (person funnel)

A standard PostHog funnel, aggregated by person (portal admin):
`createApplication.viewed` → `createApplication.selected-type` → `createApplication.created`.
Break down by `wizard_version`. This is the drop-off metric; all three steps are person-scoped
(no client exists yet at `viewed`, which is why this funnel stops at `created`).

## Insight B — Activation + time-to-integration (client_id funnel)

A PostHog funnel of `createApplication.created` → `application.first_auth`, with
**Aggregating by → `client_id`** (both events carry `client_id`). Set the funnel
**conversion window to 14 days**. This yields:
- **Activation rate** = funnel conversion %.
- **Time-to-integration** = funnel's median/p90 conversion time.
Break down by the `created` step's `application_type` (keeps M2M separate) and `wizard_version`.

If a property-aggregated funnel is unavailable, the same result via HogQL:

```sql
SELECT
  created.wizard_version,
  count() AS created_count,
  countIf(fa.first_auth_ts IS NOT NULL
          AND fa.first_auth_ts <= created.ts + INTERVAL 14 DAY) AS activated_count,
  round(100.0 * activated_count / created_count, 1) AS activation_rate_pct,
  median(dateDiff('hour', created.ts, fa.first_auth_ts)) AS ttfa_p50_hours,
  quantile(0.9)(dateDiff('hour', created.ts, fa.first_auth_ts)) AS ttfa_p90_hours
FROM (
  SELECT properties.client_id AS client_id,
         properties.wizard_version AS wizard_version,
         min(timestamp) AS ts
  FROM events WHERE event = 'createApplication.created' GROUP BY client_id, wizard_version
) AS created
LEFT JOIN (
  SELECT properties.client_id AS client_id, min(timestamp) AS first_auth_ts
  FROM events WHERE event = 'application.first_auth' GROUP BY client_id
) AS fa ON fa.client_id = created.client_id
GROUP BY created.wizard_version;
```

## Caveats

- **Emitted in real time.** `application.first_auth` is sent by an auth-server event sink the moment
  a client first authenticates (`user.authenticated` / `m2m.token.created`), so there is no cron lag.
  Its `timestamp` is the real auth time, so time-to-integration and the 14-day window are accurate.
  Baseline accrues forward from PR-1 ship — no retroactive analysis.
- **M2M defines "auth" differently** (`m2m.token.created`, a client-credentials grant). Keep it on
  its own line via `application_type`; never blend it into the interactive number.
- **Internal/test apps** inflate creation and deflate activation — exclude known internal `app_id`s.
- **Not an A/B test.** Legacy vs framework_first is time-separated; results are directional.
- **Raw first_auth counts vs. min().** The sink dedups per client with a Redis key
  (`app:<app_id>:posthog-first-auth:<client_id>`, 90-day TTL); a re-emit is only possible if that
  key is lost within the window, and the deterministic uuid still collapses it in PostHog. Always
  aggregate with `min(timestamp)` / funnel first-match (as Insight B does) — never a naïve `count()`
  of raw `application.first_auth` events.
