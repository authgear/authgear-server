---
name: write-e2e-test
description: Write end-to-end (e2e) tests for authgear-server. Use when the user asks to write, add, or create e2e tests. The tests live in e2e/tests/ and are YAML-driven.
---

Follow this guide when writing e2e tests.

## E2E Environment Setup

**CRITICAL: All `make` commands must be run from `e2e/` directory, not project root.**

### Standard Workflow

When starting fresh or after code changes:

```bash
cd e2e                          # ← MUST be in e2e directory
make teardown && make setup     # ← Tears down old containers, rebuilds binaries, applies migrations
```

This command:
- Stops and removes Docker containers (Redis, PostgreSQL, Deno hook server)
- Rebuilds authgear and authgear-portal binaries with latest code
- Applies all database migrations
- Starts services fresh

### Running Tests

After setup completes, run tests from the `e2e` directory:

```bash
cd e2e  # ← Must be in e2e directory

# Run a specific test
go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run "TestAuthflow/<folder>/<filename_without_extension>"

# Run all tests
go test ./pkg/testrunner/ -count 1 -v -timeout 10m
```

### When Setup Fails

If `make teardown && make setup` fails:

1. **Try again** - transient Docker issues are common:
   ```bash
   cd e2e && make teardown && make setup
   ```

2. **Check if containers are stuck**:
   ```bash
   docker ps  # Run from any directory to list containers
   ```

3. **If still failing** - ask the user to manually intervene. Do NOT attempt low-level Docker operations (`docker-compose down -v`, etc.) as these can affect other projects and data.

**Always use the Makefile (make teardown && make setup).** Do not skip to `./run.sh` or other workarounds.

### When to Re-setup

Always run `make teardown && make setup` from `e2e/` directory when:
- You modified server code (rebuilds binaries)
- Database schema/migrations changed
- Tests fail with DB schema errors
- Starting fresh after a long break

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

Then run the new test(s):

```
cd e2e && go test ./pkg/testrunner/ -count 1 -v -timeout 10m -run "TestAuthflow/<folder>/<filename_without_extension>"
```

If the authgear/e2e daemons are started by `./run.sh setup` and do not survive shell exit, combine setup and test in one shell:

```
./run.sh teardown
./run.sh setup
go test ./pkg/testrunner -count 1 -v -timeout 10m -run "TestAuthflow/<folder>/<filename_without_extension>"
```

If a test fails, read the error output, fix the test file, and re-run. Do not report the tests as done until they pass.

## Common Mistakes

### ❌ Testing Only Endpoint Existence, Not Functionality

**Wrong approach — just verify the endpoint works:**
```yaml
steps:
  - name: query_endpoint
    action: admin_api_graphql
    admin_api_request:
      query: |
        query { user { accountLockout { isLocked } } }
    admin_api_output:
      result: |
        {
          "data": {
            "node": {
              "accountLockout": {
                "isLocked": "[[boolean]]"
              }
            }
          }
        }
```

This test only verifies the endpoint exists and returns data. It doesn't test whether the feature actually works.

**Right approach — trigger real behavior, then verify state changed:**
```yaml
steps:
  # 1. Trigger actual failed authentication attempts
  - name: failed_login_1
    action: input
    input: |
      {
        "authentication": "primary_password",
        "password": "wrong"
      }
    output:
      error: |
        {
          "name": "InvalidCredentials"
        }

  - name: failed_login_2
    action: input
    input: |
      {
        "authentication": "primary_password",
        "password": "wrong"
      }
    output:
      error: |
        {
          "reason": "TooManyRequest"
        }

  # 2. Now verify the feature actually changed state
  - name: verify_locked
    action: admin_api_graphql
    admin_api_request:
      query: |
        query { user { accountLockout { isLocked } } }
    admin_api_output:
      result: |
        {
          "data": {
            "node": {
              "accountLockout": {
                "isLocked": true
              }
            }
          }
        }
```

The right approach:
1. **Triggers actual functionality** via auth flow `create`/`input` actions
2. **Verifies state changed** via queries and mutations
3. **Tests the feature end-to-end**, not just the endpoint

### ❌ Using Go Tests Instead of YAML Tests

**Wrong approach** — writing e2e tests in Go code:
```go
// ❌ Don't write e2e tests in Go
func TestAccountLockout(t *testing.T) {
    // Direct API calls, manual JSON marshaling, etc.
}
```

**Right approach** — use YAML-driven e2e tests:
```yaml
# ✅ Always use YAML for e2e tests
name: Account lockout prevents further attempts
steps:
  - action: input
    input: |
      {"authentication": "primary_password", "password": "wrong"}
    output:
      error: |
        {"name": "InvalidCredentials"}
```

**Why YAML?**
- Declarative, human-readable format
- Consistent with existing test suite in `e2e/tests/`
- Automatic test discovery and execution
- Built-in support for auth flows, queries, mutations, and fixtures
- Clear separation of test data from test logic

## Tips

1. Give steps descriptive `name` fields — they make failures readable and enable `{{ .steps.<name> }}` references.
2. Use comment lines to group related steps (e.g., `# Flow 1 — signup`).
3. OTP codes in test mode are always `111111` — no need to fetch them dynamically for phone/email OTP.
4. Each test case gets a fully isolated app (unique `AppID`) — no cleanup needed.
5. When testing audit events, always check the JSON path includes `payload`: `data->'payload'->'...'`.
6. Prefer `ORDER BY <stable_column>, created_at` over `ORDER BY created_at` alone to avoid flaky ordering.
7. To focus on one test during development, pass `-run "TestAuthflow/path/to/test"` to the test command.
8. After environment changes, run `make teardown && make setup` to apply latest migrations.
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

## Common Patterns

### Testing Feature Variants with Different Configurations

When a feature has multiple modes/types (e.g., per_user vs per_user_per_ip lockout), create separate test files for each variant with different authgear.yaml overrides:

```yaml
# File 1: feature_variant_a.test.yaml
authgear.yaml:
  override: |
    feature:
      mode: variant_a

# File 2: feature_variant_b.test.yaml
authgear.yaml:
  override: |
    feature:
      mode: variant_b
```

This ensures each variant is tested independently with its specific configuration.

### Testing Actual Functionality vs Just Endpoints

**Don't just verify endpoints exist** — use auth flow actions to trigger actual functionality:

**❌ Wrong**: Only query the endpoint
```yaml
steps:
  - action: admin_api_graphql
    admin_api_request:
      query: |
        query { user { lockoutStatus } }
```

**✅ Right**: Trigger functionality first, then verify state changed
```yaml
steps:
  # 1. Trigger actual failure (authentication attempt)
  - name: failed_login
    action: input
    input: |
      {"authentication": "primary_password", "password": "wrong"}
    output:
      error: |
        {"reason": "TooManyRequest"}  # Error shows it's locked

  # 2. Verify state changed via query
  - name: verify_locked
    action: admin_api_graphql
    admin_api_request:
      query: |
        query { user { accountLockout { isLocked } } }
    admin_api_output:
      result: |
        {"data": {"node": {"accountLockout": {"isLocked": true}}}}
```

### Setting Up Test Data with SQL Fixtures

Always use SQL fixtures in the `before` hooks for initial data setup, not API mutations:

```yaml
before:
  - type: custom_sql
    custom_sql:
      path: fixtures.sql  # Direct database INSERT statements

steps:
  # Now test functionality that operates on this data
  - action: admin_api_graphql
```

This is faster and more reliable than creating data via API in tests.
