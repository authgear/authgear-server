# Design

This document specifies the design of organization in Authgear.

## Data model

A new table `_auth_organization` is introduced to represent organization.

```sql
CREATE TABLE _auth_organization (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  slug text NOT NULL,
  name text NOT NULL,
)

-- Each slug is unique within the project.
CREATE UNIQUE INDEX _auth_organization_slug_unique ON _auth_organization USING btree (app_id, slug);
-- For typeahead search for slug.
CREATE INDEX _auth_organization_slug_typeahead ON _auth_organization USING btree (app_id, slug text_pattern_ops);
```

A new table `_auth_user_organization` is introduced to represent user membership in organization.

```sql
CREATE TABLE _auth_user_organization (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  updated_at timestamp without time zone NOT NULL,
  user_id text NOT NULL REFERENCES _auth_user(id),
  organization_id text NOT NULL REFERENCES _auth_organization(id)
);
-- A user and an organization can only be associated at most once.
CREATE UNIQUE INDEX _auth_user_organization_unique ON _auth_user_organization USING btree (app_id, user_id, organization_id);
-- This index supports joining from _auth_user.
CREATE INDEX _auth_user_organization_user ON _auth_user_organization USING btree (app_id, user_id);
-- This index supports joining from _auth_organization.
CREATE INDEX _auth_user_organization_organization ON _auth_user_organization USING btree (app_id, organization_id);
```

## Organization-specific configuration

Organization-specific configuration **IS NOT** stored in `_auth_organization`.
They are stored in `authgear.yaml` in the parent project.

An example looks like

```
authentication:
  identities:
  - login_id
  - oauth
  primary_authenticators:
  - password
  - passkey
  secondary_authenticators:
  - totp
  secondary_authentication_mode: if_exists
identity:
  oauth:
    providers:
    - type: azureadv2
      alias: org1_federated_login
      # NOTE: This field does not exist yet.
      # This provider is disabled in the project.
      disabled: true
authenticator:
  password:
    policy:
      min_length: 8
      uppercase_required: false
      lowercase_required: false
    expiry:
      force_change:
        enabled: false
  sms:
    phone_otp_mode: whatsapp_sms
account_deletion:
  scheduled_by_end_user_enabled: true
forgot_password:
  enabled: true
verification:
  claims:
    email:
      enabled: true
      required: true
    phone_number:
      enabled: true
      required: true

organizations:
- organization_id: xxx
  organization_specific:
    authentication:
      identities:
      # Allow Federated login only
      - oauth
      # Change MFA to required.
      secondary_authentication_mode: required
    identity:
      oauth:
        providers:
        # Enable the Federated Login OAuth provider.
        - alias: azureadv2
          disabled: false
    authenticator:
      password:
        policy:
          min_length: 12
          uppercase_required: true
          lowercase_required: true
          digit_required: true
          symbol_required: true
        expiry:
          force_change:
            enabled: true
            # Force change password every 30 days.
            duration_since_last_update: 720h
      sms:
        # Allow SMS OTP only. No Whatsapp.
        phone_otp_mode: sms
      account_deletion:
        # Disallow self-serve account deletion.
        scheduled_by_end_user_enabled: false
      forgot_password:
        # Disallow forgot password.
        # Must contact admin.
        enabled: false
      verification:
        claims:
          # The email may be fake. So verification is meaningless.
          email:
            enabled: false
            required: false
```

> [!WARNING]
> Only the above listed configuration are organization-specific.
> Other configuration are inherited to the organization.
