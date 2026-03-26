---
name: write_e2e_test
description: Write end-to-end (e2e) tests for authgear-server. Use when the user asks to write, add, or create e2e tests. The tests live in e2e/tests/ and are YAML-driven.
---

Follow this guide when writing e2e tests.

## E2E Environment

Start/restart the environment (apply latest migrations, rebuild binaries):
```
make teardown && make setup
```

If GNU make is required in your environment, use the repo's `gmake` shim:
```
PATH=/tmp/gmake-bin:$PATH ./run.sh teardown
PATH=/tmp/gmake-bin:$PATH ./run.sh setup
```

`/tmp/gmake-bin` is just a local shim directory with `make` pointing to GNU Make. Create it with:
```
mkdir -p /tmp/gmake-bin
ln -sf /usr/local/bin/gmake /tmp/gmake-bin/make
```

If `./run.sh setup` backgrounds daemons that die when the shell exits, run setup and the target test in the same shell:
```
PATH=/tmp/gmake-bin:$PATH ./run.sh teardown
PATH=/tmp/gmake-bin:$PATH ./run.sh setup
PATH=/tmp/gmake-bin:$PATH go test ./pkg/testrunner -count 1 -v -timeout 10m -run "TestAuthflow/<folder>/<filename_without_extension>"
```

Do not assume `zsh`. Run those commands sequentially in the developer's current shell.

Run a specific test:
```
cd e2e && go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run "TestAuthflow/<folder>/<filename_without_extension>"
```

Run all tests:
```
cd e2e && go test ./pkg/testrunner/ -count 1 -v -timeout 10m
```

Always run `make teardown && make setup` before running tests when:
- Migrations have changed
- Server code has changed (it rebuilds the binary)
- Tests were previously failing with DB schema errors

## Overview

E2e tests are YAML files placed under `e2e/tests/<feature>/`. The test runner auto-discovers all `*.test.yaml` files. Each file defines one test case.

## Test File Structure

```yaml
name: Human-readable test name
authgear.yaml:           # Optional: config overrides
  extend: path/to/base.yaml   # Optional base config
  override: |
    fraud_protection:
      enabled: true
before:                  # Optional: setup hooks (run before steps)
  - type: custom_sql
    custom_sql:
      path: fixtures.sql
steps:
  - name: step name      # Optional but recommended
    action: create
    input: |
      {
        "type": "signup",
        "name": "default"
      }
    output:
      result: |
        {
          "action": {
            "type": "identify"
          }
        }
```

## Step Actions

### `create` — Start a new auth flow
```yaml
- action: create
  input: |
    {
      "type": "signup",
      "name": "default"
    }
  output:
    result: |
      {
        "action": {
          "type": "identify"
        }
      }
```

### `input` — Provide input to the current flow step
```yaml
- action: input
  input: |
    {
      "identification": "phone",
      "login_id": "+6591230001"
    }
  output:
    result: |
      {
        "action": {
          "type": "verify"
        }
      }
```

To expect an error response:
```yaml
- action: input
  input: |
    {
      "channel": "sms"
    }
  output:
    error: |
      {
        "name": "TooManyRequest",
        "reason": "BlockedByFraudProtection",
        "code": 429
      }
```

### `query` — SELECT from the main app database
```yaml
- action: query
  query: |
    SELECT id, email FROM _auth_user
    WHERE app_id = '{{ .AppID }}'
    ORDER BY created_at
  query_output:
    rows: |
      [
        {
          "id": "[[string]]",
          "email": "user@example.com"
        }
      ]
```

### `audit_query` — SELECT from the audit database
```yaml
- action: audit_query
  audit_query: |
    SELECT activity_type, data->'payload'->'record'->>'decision' AS decision
    FROM _audit_log
    WHERE app_id = '{{ .AppID }}'
      AND activity_type = 'fraud_protection.decision_recorded'
    ORDER BY decision, created_at
  audit_query_output:
    rows: |
      [
        {
          "activity_type": "fraud_protection.decision_recorded",
          "decision": "allowed"
        }
      ]
```

**Important:** The `_audit_log.data` column stores the full serialized `event.Event`. JSON paths must include the `payload` level: `data->'payload'->'field'`.

**Ordering tip:** When multiple rows can land in the same second, `ORDER BY created_at` is non-deterministic. Add a secondary sort on a stable column (e.g., `ORDER BY decision, created_at`).

### `http_request` — Raw HTTP request
```yaml
- action: http_request
  http_request_method: POST
  http_request_url: http://{{ .AppID }}.authgeare2e.localhost:4000/oauth2/token
  http_request_form_urlencoded_body:
    grant_type: client_credentials
    client_id: myclient
    client_secret: mysecret
  http_output:
    http_status: 200
    json_body: |
      {
        "access_token": "[[string]]",
        "token_type": "Bearer"
      }
```

### `admin_api_graphql` — Admin API GraphQL
```yaml
- action: admin_api_graphql
  admin_api_request:
    query: |
      mutation { deleteUser(input: {userID: $userID}) { deletedUserID } }
    variables: |
      {"userID": "{{ nodeID "User" .steps.get_user.result.rows 0 .id }}"}
  admin_api_output:
    result: |
      {
        "data": {
          "deleteUser": {
            "deletedUserID": "[[string]]"
          }
        }
      }
```

### `sleep` — Wait for async operations
```yaml
- action: sleep
  sleep_for: 2s
```

## Template Variables

Available in all string fields (inputs, queries, outputs, SQL files):

| Variable | Value |
|---|---|
| `{{ .AppID }}` | Unique 32-char hex ID for this test run |
| `{{ .prev }}` | Result of the previous step |
| `{{ .steps.<name> }}` | Result of a named step |

Template functions available: all [Sprig functions](http://masterminds.github.io/sprig/) plus:
- `{{ linkOTPCode "phone" "+6591230001" }}` — Get OTP code (test mode sends `111111`)
- `{{ generateTOTPCode "secret" }}` — Generate TOTP code
- `{{ nodeID "User" "some-uuid" }}` — Encode a relay global ID
- `{{ printf "%s-suffix" .AppID }}` — String formatting

## Matcher Patterns

Use these in `output.result`, `output.error`, `query_output.rows`, and `audit_query_output.rows`:

| Pattern | Meaning |
|---|---|
| `[[string]]` | Any string value |
| `[[number]]` | Any number |
| `[[boolean]]` | Any boolean |
| `[[object]]` | Any object |
| `[[array]]` | Any array |
| `[[null]]` | Must be null |
| `[[ignore]]` | Skip this field |
| `[[never]]` | Field must not exist |
| `[["[[arrayof]]", "[[object]]"]]` | Array of any length containing objects |

Extra fields in objects are allowed by default (partial matching).

## Before Hooks

```yaml
before:
  # Run SQL fixtures against the main database
  - type: custom_sql
    custom_sql:
      path: fixtures.sql   # relative to the test file

  # Run SQL fixtures against the audit database
  - type: custom_audit_sql
    custom_audit_sql:
      path: audit_fixtures.sql

  # Import users from JSON
  - type: user_import
    user_import: users.json
```

SQL fixture files also support template variables (`{{ .AppID }}`, `{{ uuidv4 }}`, etc.).

## authgear.yaml Overrides

The `override` snippet is merged into the default config. Use it to enable features, add identity providers, or change policies:

```yaml
authgear.yaml:
  override: |
    fraud_protection:
      enabled: true
      decision:
        action: deny_if_any_warning
```

To base the test on a different config file:
```yaml
authgear.yaml:
  extend: ../base/authgear.yaml
  override: |
    authentication:
      primary_authenticators:
        - password
```

## After Writing Tests

After creating the test file(s), always run them to verify they pass.

If the e2e environment may be stale (e.g. first run in this session, or migrations/server code changed), set it up first:

```
make teardown && make setup
```

If plain `make` fails in this repo, use:

```
PATH=/tmp/gmake-bin:$PATH ./run.sh teardown
PATH=/tmp/gmake-bin:$PATH ./run.sh setup
```

If `/tmp/gmake-bin` does not exist yet, create it first:

```
mkdir -p /tmp/gmake-bin
ln -sf /usr/local/bin/gmake /tmp/gmake-bin/make
```

Then run the new test(s):

```
cd e2e && go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run "TestAuthflow/<folder>/<filename_without_extension>"
```

If the authgear/e2e daemons are started by `./run.sh setup` and do not survive shell exit, combine setup and test in one shell:

```
PATH=/tmp/gmake-bin:$PATH ./run.sh teardown
PATH=/tmp/gmake-bin:$PATH ./run.sh setup
PATH=/tmp/gmake-bin:$PATH go test ./pkg/testrunner -count 1 -v -timeout 10m -run "TestAuthflow/<folder>/<filename_without_extension>"
```

If a test fails, read the error output, fix the test file, and re-run. Do not report the tests as done until they pass.

## Tips

1. Give steps descriptive `name` fields — they make failures readable and enable `{{ .steps.<name> }}` references.
2. Use comment lines to group related steps (e.g., `# Flow 1 — signup`).
3. OTP codes in test mode are always `111111` — no need to fetch them dynamically for phone/email OTP.
4. Each test case gets a fully isolated app (unique `AppID`) — no cleanup needed.
5. When testing audit events, always check the JSON path includes `payload`: `data->'payload'->'...'`.
6. Prefer `ORDER BY <stable_column>, created_at` over `ORDER BY created_at` alone to avoid flaky ordering.
7. To focus on one test during development, pass `-run "TestAuthflow/path/to/test"` to the test command.
8. After environment changes, run `make teardown && make setup` to apply latest migrations. If `make` is broken in the current environment, use the `gmake` shim via `PATH=/tmp/gmake-bin:$PATH`.
9. Always write JSON values in `input`, `output`, `query_output`, and `audit_query_output` as multi-line for readability. Prefer:
   ```yaml
   input: |
     {
       "identification": "phone",
       "login_id": "+6591230001"
     }
   ```
   over:
   ```yaml
   input: |
     {"identification": "phone", "login_id": "+6591230001"}
   ```
10. For messaging tests in e2e, check `e2e/var/authgear.features.yaml` first. SMS and WhatsApp may be suppressed in test mode, so the right audit signals can be `sms.suppressed` / `whatsapp.suppressed` instead of `sms.sent` / `whatsapp.sent`.
11. When asserting message delivery behavior, prefer querying all relevant send/suppress audit rows for the app, normalize the target in SQL, and assert the full result set so unexpected-target rows fail the test.
12. Do not hard-code `zsh` or another shell in examples unless the task specifically depends on it.
13. Do not assume a specific interactive shell. If setup and test must run in one shell session, execute the commands sequentially in the developer's current shell.
