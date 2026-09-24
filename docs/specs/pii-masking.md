# PII Masking

- [Overview](#overview)
- [Use Cases](#use-cases)
  - [UC1: Customer support staff should not see raw contact details](#uc1-customer-support-staff-should-not-see-raw-contact-details)
  - [UC2: GDPR data minimisation — mask contact details not needed for support operations](#uc2-gdpr-data-minimisation--mask-contact-details-not-needed-for-support-operations)
  - [UC3: National ID numbers collected for KYC should not be visible to support staff](#uc3-national-id-numbers-collected-for-kyc-should-not-be-visible-to-support-staff)
  - [UC4: Audit log export pipeline must not contain raw PII](#uc4-audit-log-export-pipeline-must-not-contain-raw-pii)
  - [UC5: A trusted integration needs raw data while other consumers receive masked data](#uc5-a-trusted-integration-needs-raw-data-while-other-consumers-receive-masked-data)
  - [UC6: Support staff can look up a specific account by email but cannot browse accounts by email](#uc6-support-staff-can-look-up-a-specific-account-by-email-but-cannot-browse-accounts-by-email)
- [PII Types](#pii-types)
  - [Masking Logic](#masking-logic)
- [Configuration](#configuration)
  - [pii.masking](#piimasking)
  - [pii_type on user profile attributes](#pii_type-on-user-profile-attributes)
  - [Admin API Access Token Scopes](#admin-api-access-token-scopes)
- [Masking Behaviour](#masking-behaviour)
  - [User Profile Data](#user-profile-data)
  - [Search and Filter Requests](#search-and-filter-requests)
  - [Audit Log Data](#audit-log-data)
  - [Audit Log PII Fields](#audit-log-pii-fields)
- [Access Control Interaction](#access-control-interaction)
- [Portal Collaborator Roles](#portal-collaborator-roles)
- [Future Work](#future-work)
  - [Configurable default scopes for the Support role](#configurable-default-scopes-for-the-support-role)
  - [Per-field search permission (not supported)](#per-field-search-permission-not-supported)
  - [PII Retention config](#pii-retention-config)

## Overview

PII masking allows project owners to enforce server-side masking of personally identifiable information (PII) in API responses. When masking is enabled for an API, affected fields are replaced with a redacted representation instead of their raw value.

This is distinct from [access control](../specs/user-profile/design.md), which governs **visibility** (whether a field is returned at all). Masking operates on fields that are already visible to the caller — it changes **how** the value is presented, not **whether** it is returned.

## Use Cases

### UC1: Customer support staff should not see raw contact details

A company operates a customer support team that uses the Admin Portal to look up users and manage their accounts. Support staff do not need to see raw email addresses or phone numbers — they communicate with customers through a separate ticketing system. Support staff access the Admin API using scoped access tokens without the `pii:read` scope, so the server masks PII using the built-in defaults. Support staff see `joh****@example.com` instead of the real email, reducing unnecessary PII exposure and limiting liability in case of a support account breach.

No config change is needed — the default `pii_types` list covers all PII types.

---

### UC2: GDPR data minimisation — mask contact details not needed for support operations

Under GDPR Article 5(1)(c), personal data must be limited to what is necessary for the purpose it is processed. A company's support team communicates with users exclusively through a ticketing system (e.g. Zendesk), which already owns the email and phone channel. Support staff therefore have no legitimate purpose for seeing raw contact details in the Admin Portal — they only need the user's name to identify the case. The project owner masks email and phone while leaving names visible, satisfying the data minimisation principle without impeding support operations.

**Config:**

```yaml
pii:
  masking:
    admin_api:
      pii_types:
        - email
        - phone_number
```

Scoped access tokens issued to support staff do not include `pii:read`.

---

### UC3: National ID numbers collected for KYC should not be visible to support staff

A fintech company collects national ID numbers during KYC verification and stores them as a custom attribute. Support staff can look up accounts but should not have access to raw ID numbers — only compliance officers should. The project owner marks the attribute as `identifier` so it is masked alongside other PII when masking is enabled.

**Config:**

```yaml
user_profile:
  custom_attributes:
    attributes:
      - id: "0001"
        pointer: /x_national_id
        type: string
        pii_type: identifier
```

Scoped access tokens issued to support staff do not include `pii:read`.

---

### UC4: Audit log export pipeline must not contain raw PII

A company exports audit logs to a third-party SIEM (e.g. Datadog, Splunk) for security monitoring. Their data processing agreement with the SIEM vendor prohibits sending raw personal data. The SIEM integration uses a scoped access token without `pii:read`, so all audit log entries — including email recipients in `email.sent` events, phone numbers in `sms.sent` events, and IP addresses in event context — are masked before being exported. No config change is required.

### UC5: A trusted integration needs raw data while other consumers receive masked data

A company has both a customer support team (who should see masked data) and a backend data migration script (which needs raw email addresses to send migration notifications). Scoped access tokens issued to support staff do not include `pii:read` and receive masked PII. The migration script obtains a scoped access token with `pii:read` via Client Credentials Grant, which causes the server to bypass masking for that token only.

No config change is required.

**Migration script token request:**

```
POST /oauth2/token
grant_type=client_credentials&client_id=migration-script&client_secret=...&resource=https://auth.myapp.com/_api/admin&scope=user%3Aread+pii%3Aread
```

Scoped access tokens issued to support staff do not include `pii:read` and receive masked PII as normal.

### UC6: Support staff can look up a specific account by email but cannot browse accounts by email

A company's support team handles inbound requests from customers who identify themselves by email. When a customer says "my email is `johndoe@example.com`", the support agent needs to look up that specific account. However, the company does not want support staff to be able to filter the full user list by email or name — doing so would let them enumerate accounts and build a list of customer contacts.

Support staff are issued scoped access tokens with `pii:search` but without `pii:read`. Support staff can submit a known email as a search criterion and find the matching account, but the email is still shown masked (`joh****@example.com`) in the response. They cannot browse accounts by typing a partial name or email to discover who is registered.

**Support staff token request:**

```
POST /oauth2/token
grant_type=client_credentials&client_id=support-portal&client_secret=...&resource=https://auth.myapp.com/_api/admin&scope=user%3Aread+user%3Awrite+pii%3Asearch
```

No config change is needed — the default `pii_types` list applies. Search by email is permitted (via `pii:search`), but the returned email is still masked.

---

## PII Types

Every PII field is classified by a **pii_type**. The masking format for each type is fixed by the server and is not configurable.

| pii_type        | Example raw value                        | Masked representation          |
| --------------- | ---------------------------------------- | ------------------------------ |
| `email`         | `johndoe@example.com`                    | `joh****@example.com`          |
| `phone_number`  | `+85223456789`                           | `+8522345****`                 |
| `name`          | `John Doe`                               | `Jo** Do*`                     |
| `username`      | `johndoe`                                | `joh****`                      |
| `identifier`    | `A1234567`                               | `A123****`                     |
| `ip_address`    | `192.168.1.100`                          | `192.168.*.*`                  |
| `date_of_birth` | `1990-01-15`                             | `****-01-15`                   |
| `address`       | `{"street_address": "123 Main St", ...}` | `{"street_address": "*", ...}` |

The `pii_type` classifies the semantic meaning of a field. It is used to look up the masking format and to determine whether the field should be masked for a given API.

### Masking Logic

**`email`**

Split on `@`. For the local part, preserve the first half of the characters (floor division) and replace the rest with `*`. The domain is kept as-is.

- `user@example.com` → `us**@example.com` (local: 4 chars → preserve 2)
- `johndoe@example.com` → `joh****@example.com` (local: 7 chars → preserve 3)

**`phone_number`**

Parse using the phonenumbers library. Preserve the full country calling code. For the national significant number, preserve the first half of the digits (floor division) and replace the rest with `*`.

- `+85223456789` → `+8522345****` (national: 8 digits → preserve 4)

**`name`**

Split on whitespace. For each word, preserve the first half of the characters (floor division) and replace the rest with `*`. Rejoin with the original whitespace.

- `John Doe` → `Jo** Do*` (John: 4 → preserve 2; Doe: 3 → preserve 1)
- `Mary Jane Watson` → `Ma** Ja** Wa****`

**`username`**

Preserve the first half of the characters (floor division) and replace the rest with `*`. If the value is 1 character, replace entirely with `*`.

- `johndoe` → `joh****` (7 chars → preserve 3)
- `user` → `us**` (4 chars → preserve 2)
- `a` → `*`

**`identifier`**

Preserve the first half of the characters (floor division) and replace the rest with `*`. If the value is 1 character, replace entirely with `*`.

- `A1234567` → `A123****` (8 chars → preserve 4)
- `X1` → `X*` (2 chars → preserve 1)

**`ip_address`**

Parse the address to determine its version.

For **IPv4**, preserve the first two octets and replace the remaining two with `*`:

- `192.168.1.100` → `192.168.*.*`

For **IPv6**, preserve the first two groups and replace the remaining six with `*`:

- `2001:db8:85a3:0:0:8a2e:370:7334` → `2001:db8:*:*:*:*:*:*`

If the value cannot be parsed as a valid IP address, replace entirely with `*`.

**`date_of_birth`**

Expected format is `YYYY-MM-DD`. Replace the year component with `****`, preserving the month and day.

- `1990-01-15` → `****-01-15`

If the value does not match the expected format, replace entirely with `*`.

**`address`**

For a plain string value, replace with `*`.

For a structured object (e.g. the OIDC `address` claim), preserve the object structure and replace each string field value with `*`:

- Input: `{"street_address": "123 Main St", "locality": "New York", "region": "NY", "postal_code": "10001", "country": "US"}`
- Output: `{"street_address": "*", "locality": "*", "region": "*", "postal_code": "*", "country": "*"}`

## Configuration

### pii.masking

A new top-level `pii` section is added to `authgear.yaml`.

`pii.masking.admin_api.pii_types` is a list of `pii_type` values to mask for Admin API requests. Masking applies to scoped access tokens that do not carry the `pii:read` scope. Legacy keypair tokens (`typ: JWT`) always receive cleartext and are not affected by this config.

The default `pii_types` list covers all PII types:

```yaml
pii:
  masking:
    admin_api:
      pii_types:
        - email
        - phone_number
        - name
        - username
        - identifier
        - date_of_birth
        - address
        - ip_address
```

The `pii_types` list can be customised to mask only a subset:

```yaml
pii:
  masking:
    admin_api:
      pii_types:
        - email
        - phone_number
```

### pii_type on user profile attributes

A new optional `pii_type` field is added to both standard attribute and custom attribute config entries.

**Default `pii_type` for standard attributes**

Standard attributes have built-in default `pii_type` values. Project owners do not need to configure them unless they want to override or clear the default.

| Pointer               | Default `pii_type` |
| --------------------- | ------------------ |
| `/email`              | `email`            |
| `/phone_number`       | `phone_number`     |
| `/preferred_username` | `username`         |
| `/name`               | `name`             |
| `/given_name`         | `name`             |
| `/family_name`        | `name`             |
| `/middle_name`        | `name`             |
| `/nickname`           | `name`             |
| `/birthdate`          | `date_of_birth`    |
| `/address`            | `address`          |
| `/picture`            | (none)             |
| `/website`            | (none)             |
| `/profile`            | (none)             |
| `/gender`             | (none)             |
| `/zoneinfo`           | (none)             |
| `/locale`             | (none)             |

**Standard attributes example:**

```yaml
user_profile:
  standard_attributes:
    access_control:
      - pointer: /email
        pii_type: email
        access_control:
          end_user: readwrite
          bearer: readonly
          portal_ui: readwrite
      - pointer: /phone_number
        pii_type: phone_number
        access_control:
          end_user: readonly
          bearer: hidden
          portal_ui: readwrite
      - pointer: /given_name
        pii_type: name
        access_control:
          end_user: readwrite
          bearer: readonly
          portal_ui: readwrite
      - pointer: /family_name
        pii_type: name
        access_control:
          end_user: readwrite
          bearer: readonly
          portal_ui: readwrite
```

**Custom attributes example:**

```yaml
user_profile:
  custom_attributes:
    attributes:
      - id: "0001"
        pointer: /x_national_id
        type: string
        pii_type: identifier
        access_control:
          end_user: readwrite
          portal_ui: readwrite
```

`pii_type` is optional on standard attributes (the built-in default applies when absent) and on custom attributes (treated as non-PII when absent).

Valid values for `pii_type` on user profile attributes are: `email`, `phone_number`, `name`, `username`, `identifier`, `ip_address`, `date_of_birth`, `address`.

### Admin API Access Token Scopes

Admin API access tokens support scope-based permissions (see [admin-api-access-token.md](./admin-api-access-token.md)). The following scopes are relevant to PII masking:

| Scope        | Effect                                                                                                                                                 |
| ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `pii:read`   | Bypass `pii.masking` for this token — all PII fields returned in cleartext; also grants the ability to use PII fields as search/filter criteria        |
| `pii:search` | Allow using PII fields as search/filter criteria without bypassing response masking — responses still show masked values for tokens without `pii:read` |

`pii:read` is a superset of `pii:search`: a token with `pii:read` implicitly has search permission.

Scoped access tokens with neither scope are subject to both response masking and search restrictions described in [Masking Behaviour](#masking-behaviour). Legacy keypair tokens (`typ: JWT`) always receive cleartext and are not subject to `pii.masking`.

## Masking Behaviour

### User Profile Data

When the server serializes user profile data for a given API, it applies masking as follows:

1. Determine the calling API (e.g. `admin_api`).
2. If the request token is a legacy keypair token (`typ: JWT`) → return all values as-is.
3. If the request token is a scoped access token and `scope` contains `pii:read` → return all values as-is (masking bypassed for this token).
4. Look up `pii.masking.admin_api.pii_types` to get the list of masked pii_types.
5. For each attribute in the response:
   a. If the attribute has a `pii_type` and that type is in the masked list → replace the value with its masked representation.
   b. Otherwise → return the raw value.

Masking is applied **after** access control. If a field is `hidden` for the calling API, it is not returned at all; masking does not apply to hidden fields.

### Search and Filter Requests

When the server processes a search or filter request (e.g. listing users by email, looking up a user by login ID, filtering audit logs by recipient), it enforces search restrictions as follows:

1. If the request token is a legacy keypair token (`typ: JWT`) → all search criteria are allowed.
2. If the request token is a scoped access token and has `pii:read` or `pii:search` in its `scope` → all search criteria are allowed.
3. Otherwise, for each search criterion:
   a. Determine the pii_type of the field being searched.
   b. If the field's pii_type is in `pii.masking.admin_api.pii_types` → reject the request with an error.
   c. Otherwise → allow the criterion.

The `pii.masking.admin_api.pii_types` list therefore controls both what is masked in responses **and** what can be used as search input for tokens with neither `pii:read` nor `pii:search`. Fields excluded from `pii_types` are always searchable regardless of scope.

**Example:** A project owner who wants Support to search by username but not by email or name configures:

```yaml
pii:
  masking:
    admin_api:
      pii_types:
        - email
        - phone_number
        - name
        # username excluded — Support sees it cleartext and can search by it
```

Scoped access tokens without `pii:read` or `pii:search` can look up users by username but receive an error if they attempt to filter by email, phone, or name.

A project owner who wants Support to be able to look up a specific user by an email the customer provides (but still see masked values in responses) would grant `pii:search` without `pii:read`.

### Audit Log Data

Audit log data is accessed through the Admin API. The same masking rules as user profile data apply.

Audit log entries include two categories of PII:

**Category A — User profile attributes:** Fields drawn from `standard_attributes` or `custom_attributes` (e.g. `user.standard_attributes.email` in a `user.created` event). These are masked using the same rules as user profile data: the `pii_type` declared on the attribute config is matched against `pii.masking.admin_api.pii_types`.

**Category B — Non-profile PII:** Fields that are not user profile attributes but contain PII by nature (e.g. the recipient of a sent email). These fields have no corresponding entry in `user_profile.access_control`. Instead, the server determines their `pii_type` (see table below). The same masking rules apply: if the field's `pii_type` is in `pii.masking.admin_api.pii_types`, the field is masked.

### Audit Log PII Fields

The following lists all audit log events that contain PII fields, based on the events documented in [event.md](./event.md).

#### Event context fields

Every audit log event includes a `context` object. The following field in `context` contains PII:

| Field                | pii_type     |
| -------------------- | ------------ |
| `context.ip_address` | `ip_address` |

This applies to all audit log events.

#### Events containing `payload.user`

The `user` object includes `standard_attributes`. Fields in `standard_attributes` are masked based on the `pii_type` declared in `user_profile` config.

| Event                                                                        |
| ---------------------------------------------------------------------------- |
| `user.created`                                                               |
| `user.profile.updated`                                                       |
| `user.authenticated`                                                         |
| `user.reauthenticated`                                                       |
| `user.signed_out`                                                            |
| `user.session.terminated`                                                    |
| `user.anonymous.promoted` (both `payload.user` and `payload.anonymous_user`) |
| `user.disabled`                                                              |
| `user.reenabled`                                                             |
| `user.deletion_scheduled`                                                    |
| `user.deletion_unscheduled`                                                  |
| `user.deleted`                                                               |
| `user.anonymization_scheduled`                                               |
| `user.anonymization_unscheduled`                                             |
| `user.anonymized`                                                            |
| `authentication.identity.anonymous.failed`                                   |
| `authentication.identity.biometric.failed`                                   |
| `authentication.primary.password.failed`                                     |
| `authentication.primary.oob_otp_email.failed`                                |
| `authentication.primary.oob_otp_sms.failed`                                  |
| `authentication.secondary.password.failed`                                   |
| `authentication.secondary.totp.failed`                                       |
| `authentication.secondary.oob_otp_email.failed`                              |
| `authentication.secondary.oob_otp_sms.failed`                                |
| `authentication.secondary.recovery_code.failed`                              |
| `authentication.blocked`                                                     |
| `identity.email.added`                                                       |
| `identity.email.removed`                                                     |
| `identity.email.updated`                                                     |
| `identity.phone.added`                                                       |
| `identity.phone.removed`                                                     |
| `identity.phone.updated`                                                     |
| `identity.username.added`                                                    |
| `identity.username.removed`                                                  |
| `identity.username.updated`                                                  |
| `identity.oauth.connected`                                                   |
| `identity.oauth.disconnected`                                                |
| `identity.biometric.enabled`                                                 |
| `identity.biometric.disabled`                                                |
| `identity.email.verified`                                                    |
| `identity.phone.verified`                                                    |
| `identity.email.unverified`                                                  |
| `identity.phone.unverified`                                                  |

#### Events containing `payload.identity`

The following events include an `identity` object (or `old_identity` / `new_identity`). The `claims` field within the identity object contains PII with a server-determined `pii_type`:

| Event                         | Field                                                                                              | pii_type       |
| ----------------------------- | -------------------------------------------------------------------------------------------------- | -------------- |
| `identity.email.added`        | `payload.identity.claims.email`                                                                    | `email`        |
| `identity.email.removed`      | `payload.identity.claims.email`                                                                    | `email`        |
| `identity.email.updated`      | `payload.old_identity.claims.email`, `payload.new_identity.claims.email`                           | `email`        |
| `identity.phone.added`        | `payload.identity.claims.phone_number`                                                             | `phone_number` |
| `identity.phone.removed`      | `payload.identity.claims.phone_number`                                                             | `phone_number` |
| `identity.phone.updated`      | `payload.old_identity.claims.phone_number`, `payload.new_identity.claims.phone_number`             | `phone_number` |
| `identity.username.added`     | `payload.identity.claims.preferred_username`                                                       | `username`     |
| `identity.username.removed`   | `payload.identity.claims.preferred_username`                                                       | `username`     |
| `identity.username.updated`   | `payload.old_identity.claims.preferred_username`, `payload.new_identity.claims.preferred_username` | `username`     |
| `identity.oauth.connected`    | `payload.identity.claims.email`                                                                    | `email`        |
| `identity.oauth.disconnected` | `payload.identity.claims.email`                                                                    | `email`        |
| `identity.oauth.connected`    | `payload.identity.claims.name`                                                                     | `name`         |
| `identity.oauth.disconnected` | `payload.identity.claims.name`                                                                     | `name`         |
| `identity.oauth.connected`    | `payload.identity.claims.given_name`                                                               | `name`         |
| `identity.oauth.disconnected` | `payload.identity.claims.given_name`                                                               | `name`         |
| `identity.oauth.connected`    | `payload.identity.claims.family_name`                                                              | `name`         |
| `identity.oauth.disconnected` | `payload.identity.claims.family_name`                                                              | `name`         |
| `identity.email.verified`     | `payload.identity.claims.email`                                                                    | `email`        |
| `identity.email.unverified`   | `payload.identity.claims.email`                                                                    | `email`        |
| `identity.phone.verified`     | `payload.identity.claims.phone_number`                                                             | `phone_number` |
| `identity.phone.unverified`   | `payload.identity.claims.phone_number`                                                             | `phone_number` |

Note: OAuth identity claims depend on the provider. Only the commonly returned claims listed above are masked. Additional PII claims from specific providers are not masked unless added here in a future update.

#### Events with scalar PII fields

| Event                                     | Field                                    | pii_type                                                                                                  |
| ----------------------------------------- | ---------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `email.sent`                              | `payload.recipient`                      | `email`                                                                                                   |
| `email.error`                             | `payload.recipient`                      | `email`                                                                                                   |
| `email.suppressed`                        | `payload.recipient`                      | `email`                                                                                                   |
| `sms.sent`                                | `payload.recipient`                      | `phone_number`                                                                                            |
| `sms.error`                               | `payload.recipient`                      | `phone_number`                                                                                            |
| `sms.suppressed`                          | `payload.recipient`                      | `phone_number`                                                                                            |
| `whatsapp.sent`                           | `payload.recipient`                      | `phone_number`                                                                                            |
| `whatsapp.error`                          | `payload.recipient`                      | `phone_number`                                                                                            |
| `whatsapp.suppressed`                     | `payload.recipient`                      | `phone_number`                                                                                            |
| `authentication.identity.login_id.failed` | `payload.login_id`                       | determined at mask time: contains `@` → `email`; starts with `+` → `phone_number`; otherwise → `username` |
| `project.collaborator.invitation.created` | `payload.invitee_email`                  | `email`                                                                                                   |
| `project.collaborator.invitation.deleted` | `payload.invitee_email`                  | `email`                                                                                                   |
| `fraud_protection.decision_recorded`      | `payload.record.action_detail.recipient` | `phone_number` (only present when `payload.record.action` is `send_sms`)                                  |

## Access Control Interaction

Masking and access control are independent:

- **Access control** (the `access_control` field on each attribute) determines whether a field is visible at all (`hidden`, `readonly`, `readwrite`).
- **PII masking** determines whether a visible field is returned in cleartext or as a redacted value.

A field that is `hidden` for the calling API is never returned, regardless of `pii.masking`. A field that is `readonly` or `readwrite` may be masked if its `pii_type` matches the API's masking list.

The Admin API bypasses `access_control` (it uses `RoleGreatest` internally and always returns all attributes). PII masking is therefore the primary mechanism for restricting what raw PII the Admin API exposes.

## Portal Collaborator Roles

A new **Support** collaborator role is introduced alongside the existing **Owner** and **Editor** roles.

| Role    | PII visible                                       | PII search | User export | View Admin API key secret |
| ------- | ------------------------------------------------- | ---------- | ----------- | ------------------------- |
| Owner   | Yes                                               | Yes        | Yes         | Yes                       |
| Editor  | Yes                                               | Yes        | Yes         | Yes                       |
| Support | No — masked per `pii.masking.admin_api.pii_types` | No         | No          | No                        |

The portal issues scoped access tokens for Support collaborators using the reserved `client_id=portal`. This client ID is reserved for portal use and is not visible or configurable by project owners or collaborators.

Support collaborators are issued scoped access tokens with the following scopes:

- `user:read`
- `user:write`
- `role:read`
- `role:write`
- `group:read`
- `group:write`
- `resource:read`
- `resource:write`
- `fraud-protection:read`
- `audit-log:read`
- `user:import`

`pii:read`, `pii:search`, and `user:export` are intentionally excluded. Without `pii:search`, Support cannot use PII fields (email, phone, name) as search or filter criteria. They can still look up users by non-PII fields such as `username` if `username` is excluded from `pii.masking.admin_api.pii_types`.

**Admin API key secret:** The portal API must not allow Support collaborators to view or retrieve the `admin-api.auth` secret. A Support collaborator who can obtain it could mint a full-access legacy token, bypassing all PII restrictions of their role.

## Future Work

### Configurable default scopes for the Support role

Currently the scopes granted to Support collaborators are fixed. In future, project owners may need to tailor these defaults — for example, granting `pii:search` to Support by default, or restricting certain write scopes for more sensitive projects. A per-project configuration for the default scope set issued to Support collaborators could be introduced without breaking existing behaviour (tokens already issued to Support collaborators would continue to use the fixed defaults until the project owner opts in).

```yaml
# illustrative only — not yet implemented
portal:
  collaborators:
    roles:
      - type: support
        scopes:
          - user:read
          - user:write
          - role:read
          - role:write
          - group:read
          - group:write
          - resource:read
          - resource:write
          - fraud-protection:read
          - audit-log:read
          - user:import
          - pii:search # opted in — support can look up users by email
```

### Per-field search permission (not supported)

Currently `pii:search` is an all-or-nothing token-level scope: a token with `pii:search` may use **any** PII field as a search or filter criterion. There is no way to grant search permission for one PII field while blocking it for another.

This means the following use case is **not supported**: a project owner whose support team handles inbound tickets identified by email (and therefore needs to search by email) but does not want support staff to search by phone number — while keeping both fields masked in responses. With `pii:search`, both email and phone become searchable simultaneously.

A future `pii:search` extension could allow per-field search grants, for example:

```yaml
# illustrative only — not yet implemented
scope: "user:read pii:search:email"
```

Until then, project owners with this requirement can work around it by removing the desired search field from `pii.masking.admin_api.pii_types` (which permits searching but also unmasks the field in responses), or by accepting that `pii:search` grants broader search access than strictly needed.

### PII Retention config

The `pii` root key is designed to be extensible. Future concerns such as data retention can be added as sibling keys alongside `pii.masking` without restructuring the config:

```yaml
pii:
  masking:
    admin_api:
      pii_types:
        - email
        - phone_number
  retention: # future, illustrative only
    # ...
```
