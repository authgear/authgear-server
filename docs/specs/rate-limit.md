# Rate Limits

## Background

Authgear enforces various rate limits to address potential threats:

- Excessive resource consumption
- Brute-force attempts

## Algorithm

Rate limit in Authgear uses a variant of token bucket rate limiting,
configured by 2 variables:

- Period: The minimum period between operations
- Burst: Number of operations before additional operations is denied

When rate limit is requested, a token is taken from bucket. Buckets are filled
with burst tokens initially. Rate limit is exceeded if tokens are exhausted.
The bucket is re-filled fully after the period is elapsed from the time first
token is taken.

For rate limit on credential verification (e.g. verify password, verify OTP),
token would be taken from bucket only for failed attempts. Bucket would still
be checked to ensure tokens are available to be taken before performing
verification.
Therefore, verifying a correct password would never exceed the rate limit.

## Rate Limits

Rate limits are checked right-to-left, with short-circuit on failure.

Some considerations for rate limits design:

- Per-IP rate limit may need to be higher, due to shared IP across public WiFi users.
- Per-user rate limit before authentication may cause DoS on actual user.

Rate limits without default are hard-coded (non-configurable).

| Group                                     | Name                                           | Operation                               | Rate                                                 | Rationales                                                                                                            |
| ----------------------------------------- | ---------------------------------------------- | --------------------------------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| **authentication.general**                | `authentication.general.per_ip`                | Verify any credentials                  | 60/minute                                            | Mitigate credential brute-forcing. Per-user rate limit is not used to avoid DoS on actual user login.                 |
|                                           | `authentication.general.per_user_per_ip`       |                                         | 10/minute                                            |                                                                                                                       |
| **authentication.password**               | `authentication.password.per_ip`               | Verify Password / Additional PW         | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.password.per_user_per_ip`      |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.oob_otp.email.trigger**  | `authentication.oob_otp.email.trigger.per_ip`  | Send Email OTP                          | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **authentication.oob_otp.email.validate** | `authentication.oob_otp.email.validate.per_ip` | Verify Email OTP                        | Fallback to `authentication.general.per_ip`          | Mitigate credential brute-force                                                                                       |
|                                           |                                                |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.oob_otp.sms.trigger**    | `authentication.oob_otp.sms.trigger.per_ip`    | Send SMS OTP                            | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **authentication.oob_otp.sms.validate**   | `authentication.oob_otp.sms.validate.per_ip`   | Verify SMS OTP                          | Fallback to `authentication.general.per_ip`          | Mitigate credential brute-force                                                                                       |
|                                           |                                                |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.totp**                   | `authentication.totp.per_ip`                   | Verify TOTP                             | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.totp.per_user_per_ip`          |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.recovery_code**          | `authentication.recovery_code.per_ip`          | Verify MFA recovery code                | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.recovery_code.per_user_per_ip` |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.device_token**           | `authentication.device_token.per_ip`           | Verify MFA device tokens                | Fallback to `authentication.general.per_ip`          |                                                                                                                       |
|                                           | `authentication.device_token.per_user_per_ip`  |                                         | Fallback to `authentication.general.per_user_per_ip` |                                                                                                                       |
| **authentication.passkey**                | `authentication.passkey.per_ip`                | Verify Passkey                          | Fallback to `authentication.general.per_ip`          | Since Authgear uses discoverable credentials, user is derived from passkey and per-user-per-IP rate limit is N/A.     |
| **authentication.siwe**                   | `authentication.siwe.per_ip`                   | SWIE nonce request                      | Fallback to `authentication.general.per_ip`          | Mitigate credential brute-force. User is not known at this point for Web3 login so only per-IP rate limit is used.    |
| **authentication.signup**                 | `authentication.signup.per_ip`                 | Signup new user                         | 10/minute                                            | Mitigate resource exhaustion by rapid registration of new user.                                                       |
| **authentication.signup_anonymous**       | `authentication.signup_anonymous.per_ip`       | Signup new anonymous user               | 60/minute                                            | A more generous limit is given for anonymous user signup, since it usually occurs on app startup of new installation. |
| **authentication.account_enumeration**    | `authentication.account_enumeration.per_ip`    | Check login ID existence                | 10/minute                                            | Mitigate account enumeration.                                                                                         |
| **Features**                              |                                                |                                         |                                                      |                                                                                                                       |
| -                                         | -                                              | Login OTP failed verify attempts        | 5 attempts (per-target)                              | Mitigate credential brute-force; revoke OTP when set limit exceeded.                                                  |
| -                                         | -                                              | Verification OTP failed verify attempts | 5 attempts (per-target)                              |                                                                                                                       |
| **verification.email.trigger**            | `verification.email.trigger.per_ip`            | Send Verification Email                 | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
|                                           | `verification.email.trigger.per_user`          |                                         | Disabled                                             |                                                                                                                       |
| **verification.email.validate**           | `verification.email.validate.per_ip`           | Verify Verification Email               | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
| **verification.sms.trigger**              | `verification.sms.trigger.per_ip`              | Send Verification SMS                   | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
|                                           | `verification.sms.trigger.per_user`            |                                         | Disabled                                             |                                                                                                                       |
| **verification.sms.validate**             | `verification.sms.validate.per_ip`             | Verify Verification SMS                 | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
| **forgot_password.email.trigger**         | `forgot_password.email.trigger.per_ip`         | Send Forgot Password Email              | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **forgot_password.email.validate**        | `forgot_password.email.validate.per_ip`        | Verify Forgot Password Email            | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
| **forgot_password.sms.trigger**           | `forgot_password.sms.trigger.per_ip`           | Send Forgot Password SMS                | Disabled                                             | Mitigate mass/targeted message spam.                                                                                  |
| **forgot_password.sms.validate**          | `forgot_password.sms.validate.per_ip`          | Verify Forgot Password SMS              | 60/minute                                            | Mitigate credential brute-force.                                                                                      |
|                                           |                                                | Presign upload image request            | fixed: 10/hour                                       | Configuration not needed for now.                                                                                     |
| **oauth.register**                        | `oauth.register.per_ip`                        | Register an OAuth client via DCR        | 10/minute                                            | Mitigate resource exhaustion by rapid client registration; under open registration the endpoint is unauthenticated and writes a record per call. Mirrors `authentication.signup.per_ip`. Project-configurable — see [dcr.md](./dcr.md#rate-limits). |
|                                           | `oauth.register.per_project`                   |                                         | 1000/hour                                            | Bounds how fast a project's DCR client population can grow regardless of source IP. Not a substitute for the `oauth_client_dcr` usage limit. Project-configurable — an integration where many distinct users each self-register once (e.g. MCP-style clients) may need a higher allowance than the default; see [dcr.md](./dcr.md#rate-limits). |
| **admin_api.mutation**                    | `admin_api.mutation.all.per_project`           | Any Admin API GraphQL mutation          | 1000/minute                                          | A loose backstop against runaway or abusive object creation through the Admin API — `createGroup`, `createRole`, `createUser` and the like each write a row per call and have no cap of their own. Deliberately far above any legitimate integration; tiers are expected to set it lower. Not project-configurable; set per plan tier in `authgear.features.yaml`. See [Feature Config Rate Limits](#feature-config-rate-limits). |

| Group               | Name                         | Operation                        | Rate       | Notes                                                                                                                |
| ------------------- | ---------------------------- | -------------------------------- | ---------- | -------------------------------------------------------------------------------------------------------------------- |
|                     |                              | Send SMS (Global) (per_ip)       | Disabled   | Disabled by default; configured using environment variables                                                          |
|                     |                              | Send SMS (Global) (per_target)   | 50/day     | Configured using environment variables                                                                               |
|                     |                              | Send Email (Global) (per_ip)     | Disabled   | Disabled by default; configured using environment variables                                                          |
|                     |                              | Send Email (Global) (per_target) | 50/day     |                                                                                                                      |
| **messaging.sms**   | `messaging.sms.per_ip`       | Send SMS (Per Tenant)            | 60/minute  | Server operator can configured hard-limit using environment variable; tenant admin may set an additional rate limit. |
|                     | `messaging.sms.per_target`   | Send SMS (Per Tenant)            | 10/hour    |                                                                                                                      |
| **messaging.email** | `messaging.email.per_ip`     | Send Email (Per Tenant)          | 200/minute |                                                                                                                      |
|                     | `messaging.email.per_target` | Send Email (Per Tenant)          | 50/day     |                                                                                                                      |

## Cooldowns

Cooldowns are special rate limits, which always allow only 1 operations in a specific inverval.

Existing cooldowns are listed below:

| Name                                            | Operation                         | per-target | Rationales                           |
| ----------------------------------------------- | --------------------------------- | ---------- | ------------------------------------ |
| `authentication.oob_otp.email.trigger.cooldown` | Send authentication OOB OTP email | 1 minute   | Mitigate mass/targeted message spam. |
| `authentication.oob_otp.sms.trigger.cooldown`   | Send authentication OOB OTP SMS   | 1 minute   | Mitigate mass/targeted message spam. |
| `verification.email.trigger.cooldown`           | Send verification email           | 1 minute   | Mitigate mass/targeted message spam. |
| `verification.sms.trigger.cooldown`             | Send verification SMS             | 1 minute   | Mitigate mass/targeted message spam. |
| `forgot_password.email.trigger.cooldown`        | Send forgot password email        | 1 minute   | Mitigate mass/targeted message spam. |
| `forgot_password.sms.trigger.cooldown`          | Send forgot password SMS          | 1 minute   | Mitigate mass/targeted message spam. |

## Fallbacks

Some rate limits uses another rate limit config as a fallback if it is not set. See the following table for the mapping.

| Name                                                    | Fallback To                              |
| ------------------------------------------------------- | ---------------------------------------- |
| `authentication.password.per_ip`                        | `authentication.general.per_ip`          |
| `authentication.password.per_user_per_ip`               | `authentication.general.per_user_per_ip` |
| `authentication.oob_otp.email.validate.per_ip`          | `authentication.general.per_ip`          |
| `authentication.oob_otp.email.validate.per_user_per_ip` | `authentication.general.per_user_per_ip` |
| `authentication.oob_otp.sms.validate.per_ip`            | `authentication.general.per_ip`          |
| `authentication.oob_otp.sms.validate.per_user_per_ip`   | `authentication.general.per_user_per_ip` |
| `authentication.totp.per_ip`                            | `authentication.general.per_ip`          |
| `authentication.totp.per_user_per_ip`                   | `authentication.general.per_user_per_ip` |
| `authentication.recovery_code.per_ip`                   | `authentication.general.per_ip`          |
| `authentication.recovery_code.per_user_per_ip`          | `authentication.general.per_user_per_ip` |
| `authentication.device_token.per_ip`                    | `authentication.general.per_ip`          |
| `authentication.device_token.per_user_per_ip`           | `authentication.general.per_user_per_ip` |
| `authentication.passkey.per_ip`                         | `authentication.general.per_ip`          |
| `authentication.siwe.per_ip`                            | `authentication.general.per_ip`          |

Rate limits not mentioned in the table has no fallback.

## Feature Config Rate Limits

Most rate limits above are configured per project in `authgear.yaml`. A few are
configured in `authgear.features.yaml` instead, which a tenant admin cannot
edit — its layers (code default ← cluster ← plan ← app override) are all
operator-owned. Feature config plays two different roles here, and they are not
interchangeable:

- **Ceiling.** `messaging.rate_limits.*` caps what a project may set for itself.
  The effective bucket is whichever of the project value and the feature value
  has the lower rate (`resolveConfig`, `pkg/lib/ratelimit/ratelimits.go`), so a
  project can tighten its own messaging limits but never loosen them past the
  plan's ceiling.
- **Sole source.** `oauth.client_id_metadata_document.rate_limits.fetch.*` and
  `admin_api.rate_limits.mutation.*` have no `authgear.yaml` counterpart at all.
  There is nothing for the project to set and nothing to reconcile.

In both roles the feature config path drops the `rate_limits` segment to form
the rate limit name: `admin_api.rate_limits.mutation.all.per_project` is the
config path, `admin_api.mutation.all.per_project` is the name that appears in
the table above, in the error details, and in the audit log.

### Admin API mutations

The Admin API is the surface through which a project's own administrators (and
anything holding an Admin API key) create records: groups, roles, users,
identities, authenticators. None of those creations is individually capped, and
before this limit none was rate limited either, so a caller's write rate was
bounded only by how fast it could issue requests.

**What is counted.** One token per **top-level mutation field** of the executed
operation, taken before the operation is executed. A GraphQL document may carry
several mutation fields, and aliases allow the same field to repeat, so charging
per HTTP request would let batching buy writes at a discount. That discount is
bounded but real: `graphqlutil` already refuses a document with more than
`maxMutationFieldsPerRequest` (5) top-level mutation fields, so per-request
charging would understate the cost of a maximal document fivefold. Fields
reached through a fragment spread are counted as the fields they expand to.

Two consequences of that cap are worth stating. `n` is always between 1 and 5
when it reaches the limiter, so an `n`-token take is never large. And because
the token-bucket script only writes when the whole take conforms — and so never
improves the bucket's state when it does not — a document with more mutation
fields than `burst` can never succeed, on an empty bucket or otherwise. That is
unreachable at the default (5 ≤ 300) and only becomes reachable if a tier
configures `burst` below 5, which no tier should.

**What is not counted.** Queries are excluded. A single portal screen fires many
queries and one mutation, so a bucket sized for legitimate mutation volume would
break the console if it also had to absorb reads, and read floods are a different
problem (cost per query, not unbounded row growth) that wants a different
control. Also excluded are the Admin API's non-GraphQL endpoints, which already
have their own limits: user import/export are bounded by the `user_import_usage`
/ `user_export_usage` usage limits, and presign image upload by its fixed
10/hour.

**Buckets.** This first version gates the Admin API as a whole: one scope,
named `all`, whose bucket every mutation field consumes from. Exceeding it
fails the whole operation with
`TooManyRequest` / `RateLimited`, carrying the offending `rate_limit.name` and
`rate_limit.group` in the error details, and emits a
[`rate_limit.blocked`](./event.md#rate_limitblocked) audit log.

The rejection is **not** shaped like a GraphQL error. It happens before the
operation reaches the executor, so the response is the standard API error
envelope with HTTP 429 —

```json
{ "error": { "name": "TooManyRequest", "reason": "RateLimited", "code": 429, "info": { "rate_limit": { "name": "admin_api.mutation.all.per_project", "group": "admin_api.mutation" } } } }
```

— rather than a 200 carrying an `errors` array. This matches how the Admin API
already reports a document exceeding `maxMutationFieldsPerRequest`, but it
differs from every error raised *during* execution, so a client that only reads
`errors` will see a rate limit rejection as an empty response. Read the HTTP
status.

Authorization runs first. The Admin API's authz middleware sits ahead of this
handler in the route chain, so a request without a valid Admin API key is
refused before any token is taken — an unauthenticated caller cannot drain a
project's bucket.

| Bucket                               | Scope                  | Default     |
| ------------------------------------ | ---------------------- | ----------- |
| `admin_api.mutation.all.per_project` | Per project (`app_id`) | 1000/minute |

Three notes on that table:

- **Only per project, and deliberately so.** The Admin API authenticates as the
  project: a caller presents the project's Admin API key, and `app_id` is the
  caller identity — the one dimension the caller cannot choose. There is no
  per-IP bucket. IP would be a weaker proxy for something already measured
  directly, and it is easily sidestepped by distributing the same key across
  hosts. This is the opposite call from DCR registration and the CIMD fetch,
  which do carry per-IP buckets — but both of those endpoints are
  *unauthenticated*, so IP is the only handle they have on a caller. It is also
  a different concern from the per-IP limits on the authentication flows, which
  exist to stop credential brute-forcing; nothing here is guessable.
- **The default is a backstop, not a throttle.** 1000 mutation fields per
  minute is far above anything a portal console produces (the screens batch
  list operations, so assigning a role to twenty users is one `addRoleToUsers`
  field) and far above ordinary automation, so no legitimate integration should
  ever have to design around it. It exists to stop a runaway or hostile caller
  from writing without bound, not to shape normal traffic.
- **What it bounds, and what it does not.** It bounds the *rate*. It does not
  bound the *total*: 1000/minute sustained is a large number of rows over a
  day, and a patient caller stays under it indefinitely. Bounding the total is
  the job of a per-resource maximum, of the kind `oauth.client.maximum` and
  `collaborator.maximum` already provide and roles and groups currently lack.
  Read this limit as the burst control it is, and size tiers accordingly —
  tiers whose projects have no legitimate reason to sustain scripted Admin API
  writes are expected to set it far lower. What each tier is set to is a
  commercial decision held in the plan records, not in this repository, so no
  tier's values are recorded here. A project that genuinely needs more than its
  tier allows is a per-app feature config override by the operator, the same
  lever CIMD uses.

**Room to extend this later.** Only `mutation.all` exists today. The shape is
chosen so the rest of the vocabulary can be added without renaming it — see
[Extending the Admin API rate limits](#extending-the-admin-api-rate-limits).

**No `per_user` bucket.** The obvious third dimension is the acting user, and it
was rejected: it is only populated for portal-proxied requests (the portal passes
`actor_user_id` in the Admin API audit context), and is absent for every direct
Admin API key call — precisely the caller with the most privilege and the least
supervision. A bucket that silently does nothing for half the traffic is worse
than not having it. Worse still, where it does exist the caller controls it:
`actor_user_id` travels inside a JWT signed with the project's own Admin API
key, which tenants hold, so a direct caller could mint a fresh user id per
request and never fill the bucket.

## Future Works

- We may want to apply request-level rate limits to remaining unbounded request
  surfaces (Admin API queries, OIDC endpoints) — for the Admin API, see
  [Extending the Admin API rate limits](#extending-the-admin-api-rate-limits)
  below.
- Roles and groups have no per-project maximum, unlike OAuth clients,
  collaborators, hooks, SSO providers and NFTs. `admin_api.mutation` bounds how
  fast they can be created but not how many can exist; a maximum is the control
  that bounds the total.
- We may want to exclude certain users (e.g. by IP) from applying rate limit.

### Extending the Admin API rate limits

Only `admin_api.rate_limits.mutation.all` is implemented. The config shape
leaves room to add queries, individual fields, and the non-GraphQL endpoints
later without renaming it:

```yaml
admin_api:
  rate_limits:
    all: # every Admin API request, GraphQL and REST alike
      per_project: { ... }
    query:
      all: # every GraphQL query
        per_project: { ... }
      users: # the `users` query
        per_project: { ... }
    mutation:
      all: # every mutation field -- the only target implemented today
        per_project: { ... }
      create_group: # the `createGroup` mutation
        per_project: { ... }
```

The idea is that top-level `all` bounds Admin API requests as a whole, GraphQL
and non-GraphQL alike, while `query` and `mutation` bound GraphQL queries and
[mutation fields](#admin-api-mutations). Under those two, every child is a
scope and every child of a scope is a bucket: `all` is the reserved scope
meaning "everything under this target", and an individual field is a scope
keyed by its GraphQL field name in snake_case. Top-level `all` has no
sub-targets, so its buckets sit directly under it. How a specific scope
composes with the broader one — additively, or as a [fallback](#fallbacks) —
needs further design when it is implemented.

## Configuration

Rate limits live in one of two documents. Project-configurable ones are in
`authgear.yaml` and are covered first; the operator-owned ones are in
`authgear.features.yaml` and are covered under
[Feature config](#feature-config) below.

In general, rate limits are configured using 3 fields:

```yaml
verification:
  rate_limits:
    # No rate limit is applied
    # validate_code_per_ip:
    #   enabled: true
    #   period: 1h  # required if enabled
    #   burst: 1    # default to 1
---
verification:
  rate_limits:
    # Turn off a default rate limit
    validate_code_per_ip:
      enabled: false
---
verification:
  rate_limits:
    # 1 validation attempt allowed per hour.
    validate_code_per_ip:
      enabled: true
      period: 1h # required
      # burst: 1  # default to 1
---
verification:
  rate_limits:
    # 5 validation attempts allowed per hour.
    validate_code_per_ip:
      enabled: true
      period: 1h # required
      burst: 5 # default to 1
```

The available rate limits can be configured as follow:

```yaml
authentication:
  rate_limits:
    general:
      per_ip:
        enabled: true
        period: 1m
        burst: 60
      per_user_per_ip:
        enabled: true
        period: 1m
        burst: 10
    password:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    oob_otp:
      email:
        trigger_per_ip: # default disabled
        trigger_per_user: # default disabled
        trigger_cooldown: 1m
        max_failed_attempts_revoke_otp: # 5 # default disabled
        validate_per_ip: # default disabled
        validate_per_user_per_ip: # default disabled
      sms:
        trigger_per_ip: # default disabled
        trigger_per_user: # default disabled
        trigger_cooldown: 1m
        max_failed_attempts_revoke_otp: # 5 # default disabled
        validate_per_ip: # default disabled
        validate_per_user_per_ip: # default disabled
      whatsapp: # TBC?
    totp:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    passkey:
      per_ip: # default disabled
    siwe:
      per_ip: # default disabled
    recovery_code:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    device_token:
      per_ip: # default disabled
      per_user_per_ip: # default disabled
    signup:
      per_ip:
        enabled: true
        period: 1m
        burst: 10
    signup_anonymous:
      per_ip:
        enabled: true
        period: 1m
        burst: 60
    account_enumeration:
      per_ip:
        enabled: true
        period: 1m
        burst: 10

authenticator:
  oob_otp:
    sms:
      code_valid_period: 20m
    email:
      code_valid_period: 20m

forgot_password:
  code_valid_period: 20m
  rate_limits:
    email:
      trigger_per_ip: # default disabled
      trigger_cooldown: 1m
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60
    sms:
      trigger_per_ip: # default disabled
      trigger_cooldown: 1m
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60

verification:
  code_valid_period: 1h
  rate_limits:
    email:
      trigger_per_ip: # default disabled
      trigger_per_user: # default disabled
      trigger_cooldown: 1m
      max_failed_attempts_revoke_otp: # 5 # default disabled
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60
    sms:
      trigger_per_ip: # default disabled
      trigger_per_user: # default disabled
      trigger_cooldown: 1m
      max_failed_attempts_revoke_otp: # 5 # default disabled
      validate_per_ip:
        enabled: true
        period: 1m
        burst: 60

messaging:
  rate_limits:
    sms: # disabled
    sms_per_ip:
      enabled: true
      period: 1m
      burst: 60
    sms_per_target:
      enabled: true
      period: 1h
      burst: 10
    email: # disabled
    email_per_ip:
      enabled: true
      period: 1m
      burst: 60
    email_per_target:
      enabled: true
      period: 1h
      burst: 10
```

### Feature config

The rate limits described in
[Feature Config Rate Limits](#feature-config-rate-limits) are configured in
`authgear.features.yaml`, using the same 3 fields. The snippet below is the
**code default** — the effective config of a project on a plan that overrides
nothing:

```yaml
admin_api:
  rate_limits:
    mutation:
      # `all` is the scope covering every mutation field. Scopes for
      # individual mutations may be added later as siblings of it.
      all:
        per_project:
          enabled: true
          period: 1m
          burst: 1000

oauth:
  client_id_metadata_document:
    rate_limits:
      fetch:
        per_project:
          enabled: true
          period: 1m
          burst: 10
        per_ip:
          enabled: true
          period: 1m
          burst: 5

messaging:
  rate_limits:
    # Ceiling on what authgear.yaml's messaging.rate_limits may set.
    sms_per_ip:
      enabled: true
      period: 1m
      burst: 60
    sms_per_target:
      enabled: true
      period: 1h
      burst: 10
    email_per_ip:
      enabled: true
      period: 1m
      burst: 200
    email_per_target:
      enabled: true
      period: 24h
      burst: 50
```

A plan or app-level document states only what it changes. A tier that tightens
the Admin API mutation buckets and nothing else is written as just:

```yaml
admin_api:
  rate_limits:
    mutation:
      all:
        per_project:
          enabled: true
          period: 1m
          burst: 60
```

(Illustrative values. What each plan tier is actually set to lives in the plan
records, not here — see [Admin API mutations](#admin-api-mutations).)

Each bucket is replaced as a whole — `enabled`/`period`/`burst` are one unit, so
a layer setting `burst` alone does not inherit the lower layer's `period` — but
siblings merge independently at every level above the bucket: the document above
overrides the Admin API mutation bucket without disturbing
`oauth.client_id_metadata_document.rate_limits`, and once per-mutation scopes
exist, a document setting `create_group` alone will keep the plan's `all`.

## Audit Log

There will an audit log produced whenever any request blocked by a rate limit.

See [events](./event.md#rate_limitblocked) for details.
