- [Profile Filling](#profile-filling)
  - [Goals](#goals)
  - [Non-goals](#non-goals)
  - [Configuration reference](#configuration-reference)
    - [Form object](#form-object)
    - [New authentication flow step: fill_form](#new-authentication-flow-step-fill_form)
    - [New configuration: user_profile.forms](#new-configuration-user_profileforms)
    - [Translations](#translations)
  - [Use cases](#use-cases)
    - [UC1. Collect names at signup](#uc1-collect-names-at-signup)
    - [UC2. Ask existing users for a newly needed attribute](#uc2-ask-existing-users-for-a-newly-needed-attribute)
    - [UC3. Multi-page onboarding](#uc3-multi-page-onboarding)
    - [UC4. Validate input against an external system](#uc4-validate-input-against-an-external-system)
    - [UC5. Different applications need different attributes](#uc5-different-applications-need-different-attributes)
    - [UC6. Confirm attributes populated from an OAuth provider](#uc6-confirm-attributes-populated-from-an-oauth-provider)
    - [UC7. Collect attributes when an anonymous user is promoted](#uc7-collect-attributes-when-an-anonymous-user-is-promoted)
    - [UC8. Agree to terms and conditions at signup](#uc8-agree-to-terms-and-conditions-at-signup)
    - [UC9. Accept updated terms at login](#uc9-accept-updated-terms-at-login)
    - [UC10. Customizing the additional form order](#uc10-customizing-the-additional-form-order)
  - [The fill_form step](#the-fill_form-step)
    - [user_profile field](#user_profile-field)
      - [Fillable attributes](#fillable-attributes)
    - [Meaning of required](#meaning-of-required)
    - [When the step is shown](#when-the-step-is-shown)
    - [Pagination](#pagination)
    - [Config validation](#config-validation)
  - [Default flows](#default-flows)
  - [Where profile filling happens](#where-profile-filling-happens)
  - [Validation](#validation)
    - [Attribute constraints](#attribute-constraints)
    - [Validation hook](#validation-hook)
    - [Where validation applies](#where-validation-applies)
  - [Reading the filled attributes](#reading-the-filled-attributes)
  - [Auth UI](#auth-ui)
  - [Backward compatibility](#backward-compatibility)
  - [Future works](#future-works)
  - [Appendix: competitor review](#appendix-competitor-review)

# Profile Filling

Profile Filling lets the developer collect [Standard Attributes](../glossary.md#standard-attributes) and [Custom Attributes](../glossary.md#custom-attributes) with [forms](../glossary.md#form) in an [Authentication Flow](../glossary.md#authentication-flow). A form is one page of fields. Custom flows show it with a [`fill_form` step](#the-fill_form-step); the [default flows](#default-flows) show the forms in `user_profile.forms`.

Whether an attribute is required is a property of a form, not of the attribute. A required attribute must be filled to complete that flow. It does not guarantee that every user has the attribute: users who signed up before the form existed, users created by the Admin API, and users who cleared the attribute in the settings page may not have it.

## Goals

- Collect attributes in signup, promote, and login flows, default or custom, in Auth UI and in Custom UI.
- Prompt an existing end-user at login when an attribute required by the login flow is missing.
- Split the collected attributes across multiple pages.
- Apply the same validation in the flow, in the settings page, and in the Admin API.
- Let a hook reject input with per-attribute, localized messages.

## Non-goals

- Collecting attributes in reauth flows and account recovery flows.
- Collecting [identities](../glossary.md#identity). See [Fillable attributes](#fillable-attributes).
- Making an attribute mandatory for every user.
- Step types other than attribute forms. See [Future works](#future-works).

## Configuration reference

The configuration this spec adds, and where its rules are. Attribute definitions, including the `boolean` type and string constraints, are in [the user profile spec](./design.md).

### Form object

A Form is one page of fields. It is the `form` of a [fill_form step](#new-authentication-flow-step-fill_form), and each item of [user_profile.forms](#new-configuration-user_profileforms).

| Key                              | Type                                                      | Default  | Meaning                                                                                                                                                                                 | Rules                                             |
| -------------------------------- | --------------------------------------------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------- |
| `name`                           | string matching `^[a-zA-Z_][a-zA-Z0-9_]*$`, not `default` | none     | Names the form, so it can have its own [title and description](#translations). Forms with the same name share their title and description. | [Translations](#translations)                     |
| `show`                           | `always` or `when_incomplete`                             | `always` | When the form is shown.                                                                                                                                                                 | [When the step is shown](#when-the-step-is-shown) |
| `fields`                         | list, non-empty                                           | —        | The fields of the form, in display order. Each item has exactly one key, which names the kind of field.                                                                                 | [The fill_form step](#the-fill_form-step)         |
| `fields[].user_profile`          | object                                                    | —        | A field that reads and writes one attribute.                                                                                                                                            | [user_profile field](#user_profile-field)         |
| `fields[].user_profile.pointer`  | JSON pointer                                              | —        | The attribute. It must be [fillable](#fillable-attributes).                                                                                                                             | [Fillable attributes](#fillable-attributes)       |
| `fields[].user_profile.required` | boolean                                                   | `false`  | Whether the attribute must be filled to submit the form.                                                                                                                                | [Meaning of required](#meaning-of-required)       |

### New authentication flow step: fill_form

A step type for custom flows. It shows one [Form](#form-object) at its position in the flow.

It is available in:

- `authentication_flow.signup_flows[].steps[]`
- `authentication_flow.promote_flows[].steps[]`
- `authentication_flow.login_flows[].steps[]`
- the nested `steps` of the branches of those flows

| Key    | Type                        | Default | Meaning                                                                                                 |
| ------ | --------------------------- | ------- | ------------------------------------------------------------------------------------------------------- |
| `type` | `fill_form`                 | —       | Required. Marks the step as a form.                                                                     |
| `name` | string                      | none    | The step's name, as for any step. It is unrelated to the form's `name`.                                 |
| `form` | [Form object](#form-object) | —       | Required. The form to show.                                                                             |

```yaml
authentication_flow:
  login_flows:
    - name: default
      steps:
        - type: identify
          one_of:
            - identification: email
        - type: authenticate
          one_of:
            - authentication: primary_password
        - type: fill_form
          form:
            show: when_incomplete
            fields:
              - user_profile:
                  pointer: /x_company
                  required: true
```

Rules: [The fill_form step](#the-fill_form-step) and [Config validation](#config-validation).

### New configuration: user_profile.forms

A new object that adds Forms to the [default flows](#default-flows).

| Key                         | Type                                 | Default | Meaning                                                                       |
| --------------------------- | ------------------------------------ | ------- | ----------------------------------------------------------------------------- |
| `user_profile.forms.signup` | list of [Form objects](#form-object) | none    | Forms added to the default signup and promote flows, one page each, in order. |
| `user_profile.forms.login`  | list of [Form objects](#form-object) | none    | Forms added to the default login flow, one page each, in order.               |

```yaml
user_profile:
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /x_company
              required: true
    login:
      - show: when_incomplete
        fields:
          - user_profile:
              pointer: /x_company
              required: true
```

Rules: [Default flows](#default-flows).

### Translations

Each form page shows a title and a description above its fields:

| Key                                     | Meaning                                                                                                                                                                                                                                                            |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `v2.page.fill-form.default.title`       | The title of every form. Authgear provides it, for example "Complete your profile".                                                                                                                                                                                |
| `v2.page.fill-form.default.description` | The description of every form. Authgear provides it, worded by flow type, for example "Please provide the following information to complete your signup." in signup and promote flows, and "Please provide the following information to continue." in login flows. |
| `v2.page.fill-form.{name}.title`        | The title of the forms with that `name`. Falls back to `v2.page.fill-form.default.title`.                                                                                                                                                                          |
| `v2.page.fill-form.{name}.description`  | The description of the forms with that `name`. Falls back to `v2.page.fill-form.default.description`.                                                                                                                                                              |

The developer can override the built-in keys to change the text of every form, or set the `{name}` keys for one form. An empty translation hides the description.

These keys are used by Auth UI. The Authflow API returns the form's `name`, as `form_name`, instead of the text, so a Custom UI shows its own title and description; see [fill_form_data](../authentication-flow-api-reference.md#fill_form_data).

Fields use the existing translation keys of their attributes. Labels and `enum` option labels are defined in [Defining custom attributes](./design.md#defining-custom-attributes) and [Custom Attribute type enum](./design.md#custom-attribute-type-enum). The message for a failed `pattern` is defined in [String constraints](./design.md#string-constraints).

The hook response and the Authflow API are not configuration; see [Validation hook](#validation-hook) and [the API reference](../authentication-flow-api-reference.md#type-signup-actiontype-fill_form).

## Use cases

Each use case shows the configuration, what the end-user sees, what a Custom UI exchanges with the Authflow API, and where the developer finds the result.

Every use case except [UC5](#uc5-different-applications-need-different-attributes) and [UC10](#uc10-customizing-the-additional-form-order) uses the [default flows](#default-flows). With a custom flow, each item of `user_profile.forms.signup` or `user_profile.forms.login` becomes the `form` of a `fill_form` step, placed where the form should appear:

```yaml
# user_profile.forms.login item
- show: when_incomplete
  fields:
    - user_profile:
        pointer: /x_company
        required: true

# The same form as a step of a custom login flow
- type: fill_form
  form:
    show: when_incomplete
    fields:
      - user_profile:
          pointer: /x_company
          required: true
```

### UC1. Collect names at signup

A shopping app uses the default flows with email and password signup. It wants to greet end-users by name.

**Configuration**

```yaml
user_profile:
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /given_name
              required: true
          - user_profile:
              pointer: /family_name
              required: true
```

The default signup flow shows the form after the email and password, which the existing `authentication` and `identity` config already set up. Nothing else changes.

**What the end-user sees**

1. They enter an email address and a password, as before.
2. A page titled "Complete your profile", with the built-in description, shows two fields, Given name and Family name, both marked required. The page cannot be submitted until both are filled.
3. They enter "John" and "Doe" and select Continue. They are redirected back to the app, signed in.

**Authflow API**

After the password step, the flow responds:

```json
{
  "result": {
    "state_token": "authflowstate_1",
    "type": "signup",
    "name": "default",
    "action": {
      "type": "fill_form",
      "data": {
        "type": "fill_form_data",
        "fields": [
          {
            "key": "/given_name",
            "type": "string",
            "label": "Given name",
            "required": true
          },
          {
            "key": "/family_name",
            "type": "string",
            "label": "Family name",
            "required": true
          }
        ]
      }
    }
  }
}
```

The Custom UI renders the form and sends:

```json
{
  "state_token": "authflowstate_1",
  "input": {
    "fields": [
      { "key": "/given_name", "value": "John" },
      { "key": "/family_name", "value": "Doe" }
    ]
  }
}
```

The next response has `action.type: finished`.

Sending the input without `/family_name` fails with `ValidationFailed`, and the flow stays in this step.

**Where the attribute appears**

- User Info endpoint, called by the app with the user's access token:

  ```json
  {
    "sub": "user_id",
    "email": "john@example.com",
    "given_name": "John",
    "family_name": "Doe"
  }
  ```

- Portal: User Management > the user > Profile > Personal Information.
- Settings page: the end-user can edit both names under Profile.
- Webhook `user.profile.updated`, with `payload.user.standard_attributes.given_name` set to `"John"`.

### UC2. Ask existing users for a newly needed attribute

A B2B app starts invoicing and needs each active user's company name. It already has 10,000 users, none of them with a company.

**Configuration**

```yaml
user_profile:
  custom_attributes:
    attributes:
      - id: "0001"
        pointer: /x_company
        type: string
        max_length: 100
        access_control:
          end_user: readwrite
          bearer: readonly
          portal_ui: readwrite
  forms:
    signup:
      - name: company
        fields:
          - user_profile:
              pointer: /x_company
              required: true
    login:
      - name: company
        show: when_incomplete
        fields:
          - user_profile:
              pointer: /x_company
              required: true
```

New users fill the company at signup. Existing users are asked at login. Both forms are named `company`, so they share one title and description, in `templates/en/translation.json`:

```json
{
  "v2.page.fill-form.company.title": "Your company",
  "v2.page.fill-form.company.description": "We print your company name on invoices."
}
```

**What the end-user sees**

The login form has `show: when_incomplete`, so it is [shown only while `/x_company` is not filled](#when-the-step-is-shown). Each user sees it at most once, unless the value is cleared later. Without `show: when_incomplete`, the page would appear on every login.

| Who signs in                                                            | What they see after the password                                                                                                                                    |
| ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| An existing user without a company, on the first login after the change | A page titled "Your company", with the description "We print your company name on invoices." and one required field, Company. After filling it, they reach the app. |
| The same user on any later login                                        | Nothing. `/x_company` is filled, so the form is skipped.                                                                                                            |
| A user who signed up after the change                                   | Nothing. They filled Company at signup.                                                                                                                             |
| A user whose company was cleared, by an admin or in the settings page   | The page again, on their next login.                                                                                                                                |

A user who was already signed in when the config changed keeps using the app without interruption. Refresh tokens and `prompt=none` do not run a login flow, so they see the page the next time they sign in.

**Authflow API**

After `authenticate`, the response is `action.type: fill_form` with this data:

```json
{
  "type": "fill_form_data",
  "form_name": "company",
  "fields": [
    {
      "key": "/x_company",
      "type": "string",
      "label": "Company",
      "required": true,
      "constraints": { "max_length": 100 }
    }
  ]
}
```

The input is `{"fields": [{"key": "/x_company", "value": "Oursky"}]}`.

A user who already has a company never receives this action; the response after `authenticate` is `finished`.

**Where the attribute appears**

- User Info endpoint: `"custom_attributes": { "x_company": "Oursky" }`.
- Portal: User Management > the user > Profile > Custom Attributes.
- Until everyone has signed in again, some users have no `x_company`. The invoicing backend must handle a missing value.

### UC3. Multi-page onboarding

A fitness app asks for personal details first and preferences second. Preferences are optional.

**Configuration** (`/x_fitness_goal` and `/x_referral_source` are `enum` custom attributes with `end_user: readwrite`)

```yaml
user_profile:
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /given_name
              required: true
          - user_profile:
              pointer: /birthdate
              required: true
      - fields:
          - user_profile:
              pointer: /x_fitness_goal
              required: false
          - user_profile:
              pointer: /x_referral_source
              required: false
```

Each item of `signup` is one page.

**What the end-user sees**

1. Page 1: Given name and a date picker for Birthdate, both required.
2. Page 2: two dropdowns, Fitness goal and How did you hear about us, both optional. They can continue with both empty.

**Authflow API**

Two consecutive `fill_form` actions, each with its own `state_token`. Submitting page 2 with `{"fields": []}` is valid, because nothing on it is required.

**Where the attribute appears**

- User Info endpoint: `"birthdate": "1990-05-01"` at the root. If the end-user filled page 2, `custom_attributes` also has `x_fitness_goal`.
- An optional attribute left empty is absent from the response, not `null`.

### UC4. Validate input against an external system

An employee portal collects an employee ID. The format is checked by an [attribute constraint](#attribute-constraints). Whether the ID exists is checked by a [validation hook](#validation-hook) that calls the HR system.

**Configuration**

```yaml
user_profile:
  custom_attributes:
    attributes:
      - id: "0002"
        pointer: /x_employee_id
        type: string
        pattern: "^E[0-9]{5}$"
        access_control:
          end_user: readwrite
          bearer: readonly
          portal_ui: readwrite
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /x_employee_id
              required: true
hook:
  blocking_handlers:
    - event: user.profile.pre_update
      url: https://hr-check.example.com/hook
```

The translation `custom-attribute-pattern-error-/x_employee_id` is set to "Format: E followed by 5 digits".

**What the end-user sees**

- They type `12345`. The page shows "Format: E followed by 5 digits" under the field.
- They type `E99999`, which has the right format but doesn't exist. The hook answers in `context.language`, here `zh-HK`:

  ```json
  {
    "is_allowed": false,
    "title": "找不到員工",
    "reasons": [
      {
        "type": "invalid_profile_attribute",
        "pointer": "/x_employee_id",
        "message": "這個員工編號不存在。"
      }
    ]
  }
  ```

  The page shows "找不到員工" above the form and "這個員工編號不存在。" under the field.

- They type `E10234`, which the hook allows, and continue.

**Authflow API**

- The `12345` input fails with `ValidationFailed`. Its `info.causes` points at the value of `/x_employee_id`.
- The `E99999` input fails with:

  ```json
  {
    "error": {
      "name": "Forbidden",
      "reason": "HookDisallowed",
      "info": {
        "reasons": [
          {
            "title": "找不到員工",
            "reasons": [
              {
                "type": "invalid_form_field",
                "key": "/x_employee_id",
                "message": "這個員工編號不存在。"
              }
            ]
          }
        ]
      }
    }
  }
  ```

In both cases the flow stays in the step, and the same `state_token` can be used again.

**Where the attribute appears**

- User Info endpoint: `"custom_attributes": { "x_employee_id": "E10234" }`.
- If an admin later sets `E99999` through the Admin API, the same hook fires with `context.triggered_by: admin_api`, and the mutation fails with `HookDisallowed`. There, the hook's `reasons` are returned unchanged.

### UC5. Different applications need different attributes

One project serves a store and a forum. The store needs a shipping address; the forum needs a nickname.

This needs custom flows: each application needs its own signup and login flows, and there is only one default flow of each type.

**Configuration**

```yaml
authentication_flow:
  login_flows:
    - name: store
      steps:
        - type: identify
          one_of:
            - identification: email
        - type: authenticate
          one_of:
            - authentication: primary_password
        - type: fill_form
          form:
            show: when_incomplete
            fields:
              - user_profile:
                  pointer: /address
                  required: true
    - name: forum
      steps:
        - type: identify
          one_of:
            - identification: email
        - type: authenticate
          one_of:
            - authentication: primary_password
        - type: fill_form
          form:
            show: when_incomplete
            fields:
              - user_profile:
                  pointer: /nickname
                  required: true
  # signup_flows named store and forum carry the same steps.
ui:
  authentication_flow:
    groups:
      - name: store
        flows:
          - type: signup
            name: store
          - type: login
            name: store
      - name: forum
        flows:
          - type: signup
            name: forum
          - type: login
            name: forum
oauth:
  clients:
    - client_id: store
      x_authentication_flow_allowlist:
        groups:
          - name: store
    - client_id: forum
      x_authentication_flow_allowlist:
        groups:
          - name: forum
```

**What the end-user sees**

1. They sign up in the forum and are asked only for a nickname.
2. Later they sign in to the store with the same account. After the password, they see an address form: street, city, region, postal code, country. The forum never asked for any of it.
3. Signing in to the forum again shows no extra page, because they have a nickname.

If they already have a session from the forum, the store's continue screen shows the address form before returning to the store.

**Authflow API**

A Custom UI for the store creates the flow with `{"type": "login", "name": "store"}`. The `fill_form` action lists `/address` once. The input sends it as one object:

```json
{
  "fields": [
    {
      "key": "/address",
      "value": {
        "street_address": "1 Queen's Road",
        "locality": "Central",
        "country": "HK"
      }
    }
  ]
}
```

**Where the attribute appears**

- User Info endpoint: `"nickname"` and `"address": { ... }` at the root. Both the store and the forum see both attributes, because attributes belong to the user, not to the application.

### UC6. Confirm attributes populated from an OAuth provider

A travel app lets end-users sign up with Google. It wants a correct legal name, so the end-user reviews the name Google returned before the account is created.

**How it works**

1. The default signup flow's `identify` step creates the Google identity. Because of [population](./design.md#standard-attributes-population) with `strategy: on_signup`, that step also copies Google's `given_name` and `family_name` claims into the new user's standard attributes.
2. The next step is the form from `user_profile.forms.signup`. It uses the default [`show: always`](#when-the-step-is-shown), and each field shows the attribute's current value. So the page opens with the Google values already filled in.
3. The end-user keeps or edits the values and submits. The submitted values replace the populated ones.

If the end-user signs up with email instead, population has nothing to copy, and the same page opens with empty fields.

**Configuration**

```yaml
user_profile:
  standard_attributes:
    population:
      strategy: on_signup
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /given_name
              required: true
          - user_profile:
              pointer: /family_name
              required: true
          - user_profile:
              pointer: /gender
              required: false
```

Google is set up as an OAuth provider in `identity.oauth.providers`, and email signup is also enabled.

**What the end-user sees**

1. They select "Continue with Google" and approve on Google's page.
2. Back in Auth UI, a page shows Given name "John" and Family name "Doe", taken from Google. Gender is empty.
3. They change the family name to "Doe-Smith" and continue. They are redirected to the app, signed in.

The page appears even though both required fields are already filled, because the step uses `show: always`. With `show: when_incomplete`, it would be skipped here, because Google filled both required names.

**Authflow API**

After the OAuth callback, `data.fields` carries the populated values:

```json
[
  {
    "key": "/given_name",
    "type": "string",
    "label": "Given name",
    "required": true,
    "value": "John"
  },
  {
    "key": "/family_name",
    "type": "string",
    "label": "Family name",
    "required": true,
    "value": "Doe"
  },
  {
    "key": "/gender",
    "type": "string",
    "label": "Gender",
    "required": false,
    "suggestions": ["male", "female"]
  }
]
```

The input must send every value to keep:

```json
{
  "fields": [
    { "key": "/given_name", "value": "John" },
    { "key": "/family_name", "value": "Doe-Smith" }
  ]
}
```

A field left out of the input is treated as empty. Here `/gender` stays empty.

**Where the attribute appears**

- User Info endpoint: `"given_name": "John"`, `"family_name": "Doe-Smith"`.
- The Google identity keeps the original claims, readable through the Admin API `Identity.claims`. Editing the form does not change them.

### UC7. Collect attributes when an anonymous user is promoted

A game lets people play as anonymous users. When they create a real account, the game needs a public nickname of 3 to 16 letters or digits.

**Configuration**

```yaml
user_profile:
  standard_attributes:
    validation:
      - pointer: /nickname
        min_length: 3
        max_length: 16
        allowed_characters: [letters, digits]
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /nickname
              required: true
```

The default promote flow uses the `signup` forms, so players who sign up directly also pick a nickname. To ask only on promotion, use a custom promote flow with the `fill_form` step instead.

**What the end-user sees**

1. They choose "Save my progress" and enter an email and a password.
2. A page asks for a nickname. Under the field: "3–16 characters. Letters and digits only."
3. They type `Ace!`. The field shows "Use only letters and digits."
4. They type `Ace` and continue. Their game progress is kept, because the user ID does not change.

The same rules apply if they change the nickname later in the settings page, or if an admin sets it through the Admin API.

**Authflow API**

The promote flow returns `type: promote` with a `fill_form` action and input shaped like [UC1](#uc1-collect-names-at-signup), with one `/nickname` field. Sending `Ace!` fails with `ValidationFailed`, and the flow stays in the step.

**Where the attribute appears**

- Webhook `user.anonymous.promoted`, followed by `user.profile.updated` with `payload.user.standard_attributes.nickname`.
- User Info endpoint: `"nickname"`, and `https://authgear.com/claims/user/is_anonymous` becomes `false`.

### UC8. Agree to terms and conditions at signup

A booking app requires new users to accept its terms of service and privacy policy. It also asks for optional marketing consent.

**Configuration**

```yaml
user_profile:
  custom_attributes:
    attributes:
      - id: "0003"
        pointer: /x_accepted_terms_2026_01
        type: boolean
        access_control:
          end_user: readwrite
          bearer: readonly
          portal_ui: readonly
      - id: "0004"
        pointer: /x_marketing_consent
        type: boolean
        access_control:
          end_user: readwrite
          bearer: readonly
          portal_ui: readonly
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /x_accepted_terms_2026_01
              required: true
          - user_profile:
              pointer: /x_marketing_consent
              required: false
```

Labels, in `templates/en/translation.json`. They use the existing [label key](./design.md#defining-custom-attributes) `custom-attribute-label-{pointer}`:

```json
{
  "custom-attribute-label-/x_accepted_terms_2026_01": "I agree to the <a href=\"https://booking.example.com/terms\">Terms of Service</a> and <a href=\"https://booking.example.com/privacy\">Privacy Policy</a>",
  "custom-attribute-label-/x_marketing_consent": "Send me offers and news by email"
}
```

The pointer includes the terms version, so a later version gets a new attribute. See [UC9](#uc9-accept-updated-terms-at-login).

**What the end-user sees**

1. They enter an email address and a password.
2. A page shows two checkboxes. The first reads "I agree to the Terms of Service and Privacy Policy"; both names are links that open in a new tab. The first checkbox is marked required.
3. With the first checkbox unchecked, the page cannot be submitted.
4. They check the first box, leave the marketing box unchecked, and continue.

**Authflow API**

`data.fields`:

```json
[
  {
    "key": "/x_accepted_terms_2026_01",
    "type": "boolean",
    "label": "I agree to the <a href=\"https://booking.example.com/terms\">Terms of Service</a> and <a href=\"https://booking.example.com/privacy\">Privacy Policy</a>",
    "required": true
  },
  {
    "key": "/x_marketing_consent",
    "type": "boolean",
    "label": "Send me offers and news by email",
    "required": false
  }
]
```

Input:

```json
{
  "fields": [
    { "key": "/x_accepted_terms_2026_01", "value": true },
    { "key": "/x_marketing_consent", "value": false }
  ]
}
```

Sending `"value": false` for `/x_accepted_terms_2026_01` fails with `ValidationFailed`.

A Custom UI receives the same translated label, links included, in the field's `label`.

**Where the attribute appears**

- User Info endpoint: `"custom_attributes": { "x_accepted_terms_2026_01": true, "x_marketing_consent": false }`.
- Portal: User Management > the user > Profile > Custom Attributes, read-only because `portal_ui` is `readonly`.
- Webhook `user.profile.updated`. A backend that must keep proof of consent stores `context.timestamp` and the attribute value from this event, because Authgear keeps the current value only.
- Settings page: the end-user can uncheck either box. Unchecking the terms box does not delete the account; if `user_profile.forms.login` also has the form, as in [UC9](#uc9-accept-updated-terms-at-login), they are asked again at their next login.

### UC9. Accept updated terms at login

The booking app publishes new terms. Every user must accept them at their next login.

**Configuration**

The developer adds a `boolean` custom attribute `/x_accepted_terms_2026_07`, configured like `/x_accepted_terms_2026_01` in UC8, with a label linking to the new terms. Both forms use the new attribute:

```yaml
user_profile:
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /x_accepted_terms_2026_07
              required: true
          - user_profile:
              pointer: /x_marketing_consent
              required: false
    login:
      - show: when_incomplete
        fields:
          - user_profile:
              pointer: /x_accepted_terms_2026_07
              required: true
```

**What the end-user sees**

- **An existing user signing in:** after the password, a page with one checkbox linking to the new terms. After checking it, they reach the app.
- **A user who signed up after the change:** no extra page at login, because they accepted `/x_accepted_terms_2026_07` at signup.
- **A user who does not accept:** they cannot finish signing in. Leaving the page leaves them signed out.

**Authflow API**

After `authenticate`, users without `x_accepted_terms_2026_07: true` receive the `fill_form` action. Users who have it receive `finished`.

**Where the attribute appears**

- User Info endpoint: `custom_attributes` has both `x_accepted_terms_2026_01` and `x_accepted_terms_2026_07`. The application checks the current version's attribute.

### UC10. Customizing the additional form order

An employee app signs users up with a phone number and an SMS OTP. Each SMS costs money, so the app checks the employee ID before sending one. The default flows always put forms after the OTP, so the app uses a custom signup flow that puts the form right after the phone number is entered.

**Configuration**

`/x_employee_id` and its validation hook are configured as in [UC4](#uc4-validate-input-against-an-external-system).

```yaml
authentication_flow:
  signup_flows:
    - name: default
      steps:
        - name: setup_phone
          type: identify
          one_of:
            - identification: phone
        - type: fill_form
          form:
            fields:
              - user_profile:
                  pointer: /x_employee_id
                  required: true
        - type: create_authenticator
          one_of:
            - authentication: primary_oob_otp_sms
              target_step: setup_phone
        - type: verify
          target_step: setup_phone
```

The flow is named `default`, so it replaces the default signup flow, and `user_profile.forms.signup` no longer applies to signup. See [Default flows](#default-flows).

**What the end-user sees**

1. The first page asks for a phone number. No SMS is sent yet.
2. The next page asks for an employee ID. They enter `E99999`. The hook rejects it, and the page shows the hook's message under the field. Still no SMS is sent.
3. They enter `E10234`. The SMS OTP is sent.
4. They enter the OTP and are signed in.

The form cannot come before the `identify` step; see [Config validation](#config-validation).

**Authflow API**

After the `identify` input, the response is `action.type: fill_form`. After the form is accepted, the flow continues to `create_authenticator`, which sends the OTP. A rejected form stays in `fill_form` with the `HookDisallowed` error shown in [UC4](#uc4-validate-input-against-an-external-system).

**Where the attribute appears**

The same places as in [UC4](#uc4-validate-input-against-an-external-system). The employee ID is stored when the user is created at the end of the flow. An end-user who stops before verifying the phone number creates no user.

## The fill_form step

A `fill_form` step shows one form. Where it is available is listed in the [configuration reference](#new-authentication-flow-step-fill_form).

```yaml
- type: fill_form
  form:
    show: when_incomplete
    fields:
      - user_profile:
          pointer: /given_name
          required: true
      - user_profile:
          pointer: /x_newsletter
          required: false
```

The step's keys are listed in the [configuration reference](#new-authentication-flow-step-fill_form), and the keys of its `form` in [Form object](#form-object). The only kind of field is [`user_profile`](#user_profile-field).

The Authflow API input and output of this step are in [the API reference](../authentication-flow-api-reference.md#type-signup-actiontype-fill_form). The API describes each field by a key and a value type, not by its kind, so a frontend renders every kind the same way.

### user_profile field

A `user_profile` field reads and writes one attribute. It follows the [User Profile Pointer](../convention.md#user-profile-pointer) convention.

Its keys, `pointer` and `required`, are listed in the [configuration reference](#form-object).

A `user_profile` field is **incomplete** when it is required and its attribute is not filled. An optional `user_profile` field is never incomplete.

In the Authflow API, the field's `key` is its pointer, and its `type` is:

| Attribute                                                                                                    | `type`                                                      |
| ------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------- |
| `/name`, `/given_name`, `/family_name`, `/middle_name`, `/nickname`                                          | `string`                                                    |
| `/gender`                                                                                                    | `string`, with `suggestions` `male` and `female`            |
| `/birthdate`                                                                                                 | `date`                                                      |
| `/zoneinfo`                                                                                                  | `timezone`                                                  |
| `/locale`                                                                                                    | `enum`, with the project's supported languages as `options` |
| `/profile`, `/picture`, `/website`                                                                           | `url`                                                       |
| `/address`                                                                                                   | `address`                                                   |
| Custom attribute of type `string`, `integer`, `number`, `enum`, `phone_number`, `email`, `url`, or `boolean` | The same name                                               |
| Custom attribute of type [`alpha2`](./design.md#custom-attribute-type-alpha2)                                | `country_code`                                              |

The `label` and `enum` option labels use the attribute's existing [translations](#translations). `constraints` are the attribute's [string constraints](./design.md#string-constraints) or its `minimum` and `maximum`.

#### Fillable attributes

An attribute is fillable when all of the following hold:

| Rule                                                                                      | Reason                                                                                                                                                      |
| ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Its `end_user` [access control](./design.md#access-control-configuration) is `readwrite`. | The end-user must be able to change what they are asked for.                                                                                                |
| It is not `/email`, `/phone_number`, or `/preferred_username`.                            | They are [coupled with identities](./design.md#list-of-standard-attributes-that-are-coupled-with-identities); collect them with an `identify` step instead. |
| It is not a sub-attribute of `/address`, such as `/address/locality`.                     | `/address` is filled as a whole.                                                                                                                            |

An attribute is **filled** when it has a value, with two exceptions:

- `/address` is filled when at least one of its sub-attributes is non-empty.
- A [`boolean`](./design.md#custom-attribute-type-boolean) custom attribute is filled only when it is `true`.

Constraints are not considered. A stored value that no longer meets its attribute's constraints, because they were narrowed after it was written, is still filled.

### Meaning of required

A required `user_profile` field means the form cannot be submitted until the attribute is filled. For a `boolean` attribute, the checkbox must be checked. It has no effect outside the flow:

- Users who never ran the flow may lack the attribute.
- The Admin API can create users without it or clear it.
- The end-user can clear it in the settings page.

A login form with `show: when_incomplete` asks again at the next login. See [UC2](#uc2-ask-existing-users-for-a-newly-needed-attribute).

### When the step is shown

The rule is the same in every flow type.

| `show`             | The form is shown                                                                                        |
| ------------------ | -------------------------------------------------------------------------------------------------------- |
| `always` (default) | Every time the flow reaches it.                                                                          |
| `when_incomplete`  | Only when the form is **incomplete**: at least one of its fields is incomplete. Otherwise it is skipped. |

Each kind of field defines when it is incomplete; see [user_profile field](#user_profile-field).

The check uses the user's attributes when the flow reaches the step. In a signup or promote flow, those include values [populated](./design.md#standard-attributes-population) by an earlier `identify` step. See [UC6](#uc6-confirm-attributes-populated-from-an-oauth-provider).

When the step is shown, the form includes every field, optional and already-filled ones included. Filled fields show their current value. A current value is checked against the current constraints on submission like any other, so a value that no longer meets them must be corrected, or cleared if the field is optional, before the form can be submitted.

When a signup flow continues as an existing user through [account linking](../account-linking.md), its forms are skipped, because the flow no longer sets up a new user.

### Pagination

Each form is one page. To split a form, write consecutive `fill_form` steps, or list several forms in `user_profile.forms`. Each form applies its own `show`.

### Config validation

A config is rejected when a form:

- has a `user_profile` field whose attribute is not [fillable](#fillable-attributes);
- has two `user_profile` fields with the same pointer;
- is named `default`, because `v2.page.fill-form.default.*` are the built-in keys; or
- has `show: when_incomplete` and no field that can be incomplete, such as a form with only optional fields, because such a form is never shown.

A config is also rejected when a `fill_form` step can be reached, on any path through its flow, before the end-user is identified in a signup or promote flow, or before the end-user is authenticated in a login flow. Otherwise the form could write to the profile of a user who has not proven who they are, and the [validation hook](#validation-hook) would receive a user without identities.

The rules apply to the `form` of every `fill_form` step and to every form in [`user_profile.forms`](#default-flows). A change to access control that makes an attribute used by any form non-fillable is also rejected.

## Default flows

The default flows are the flows named `default`, which Authgear generates from the project config. To add forms after their usual signup and login steps, list them in `user_profile.forms`:

```yaml
user_profile:
  forms:
    signup:
      - fields:
          - user_profile:
              pointer: /given_name
              required: true
          - user_profile:
              pointer: /family_name
              required: true
    login:
      - show: when_incomplete
        fields:
          - user_profile:
              pointer: /given_name
              required: true
          - user_profile:
              pointer: /family_name
              required: true
```

`signup` holds the forms of the default signup and promote flows, and `login` the forms of the default login flow. See the [configuration reference](#new-configuration-user_profileforms).

`user_profile.forms` only affects the default flows, in the same way as [`bot_protection.requirements`](../bot-protection.md#behavior-of-builtin-flows). Custom flows use `fill_form` steps instead.

Each item is a [form](#form-object), and follows the same rules for [showing](#when-the-step-is-shown), [pagination](#pagination), and [config validation](#config-validation). When a key is absent, that default flow has no additional forms.

The default signup login flow delegates to the default signup and login flows, so it gets both lists.

| Default flow    | Where the forms are added                                                                               |
| --------------- | ------------------------------------------------------------------------------------------------------- |
| Signup, promote | After identification, authentication, and verification, and before the passkey prompt.                  |
| Login           | After the end-user is authenticated and their account status is checked, and before the passkey prompt. |

A custom flow named `default` replaces the default flow of its type. For that flow type, `user_profile.forms` has no effect.

Default reauth and account recovery flows never get these forms.

## Where profile filling happens

Profile filling happens only inside a flow that has a form, either a `fill_form` step or a form from `user_profile.forms`:

| Entry point                                                                         | Profile filling                                                                                                    |
| ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Signup, promote, and login flows, in Auth UI and Custom UI                          | The flow's forms.                                                                                                  |
| Continuing an existing session in Auth UI                                           | The forms of the login flow Auth UI would use for the authorization request, each applying its `show`.             |
| [`select_account`](../custom-ui-select-account.md) in Custom UI                     | The login flow's later steps, including its forms.                                                                 |
| Reauth flows, account recovery flows                                                | None.                                                                                                              |
| `prompt=none`, refresh token grant, biometric login, app2app, pre-authenticated URL | None. These do not run a login flow; the end-user is asked at their next login.                                    |

## Validation

### Attribute constraints

Values are checked against the attribute's own constraints, on every write path:

- `minimum` and `maximum` on [`integer`](./design.md#custom-attribute-type-integer) and [`number`](./design.md#custom-attribute-type-number) custom attributes;
- [string constraints](./design.md#string-constraints): `min_length`, `max_length`, `allowed_characters`, and `pattern`;
- the fixed formats of standard attributes and other custom attribute types.

Auth UI shows a failed constraint's message under the field.

### Validation hook

The blocking event [user.profile.pre_update](../event.md#userprofilepre_update) fires on every attribute write, including submission of a form. The payload contains the user with the submitted attributes applied.

To reject specific attributes, the hook returns [`reasons`](../event.md#userprofilepre_update) of type `invalid_profile_attribute`. See [UC4](#uc4-validate-input-against-an-external-system) for an example. The hook localizes its messages using `context.language`; Authgear displays them unchanged.

Each entry of `info.reasons` of the resulting `HookDisallowed` error carries the hook's `title` and `reason`, when present, and its `reasons`:

| Where                                              | `reasons` items                                                                                                                              |
| -------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Authflow API, when a form is submitted            | Each `invalid_profile_attribute` becomes `{ "type": "invalid_form_field", "key", "message" }`, where `key` is the key of the matching field. |
| Admin API and the settings page                    | As returned by the hook.                                                                                                                     |

### Where validation applies

| Write path              | Attribute constraints | `user.profile.pre_update`                              | `required`   |
| ----------------------- | --------------------- | ------------------------------------------------------ | ------------ |
| `fill_form` in any flow | ✅                    | ✅                                                     | Per step     |
| Settings page           | ✅                    | ✅                                                     | Not enforced |
| Admin API and portal    | ✅                    | ✅ (`context.triggered_by` is `admin_api` or `portal`) | Not enforced |

## Reading the filled attributes

Filled attributes are stored like any other attribute write. Other components read them through the existing channels:

| Reader                                        | Channel                                                                                                                                                                       | Condition                                                |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| The application, with the user's access token | [User Info endpoint](./design.md#user-info-endpoint). Standard attributes are at the root; custom attributes are under `custom_attributes`.                                   | The attribute's `bearer` access control is not `hidden`. |
| A backend server, per request                 | JWT access token claims added by a hook on [oidc.jwt.pre_create](../event.md#oidcjwtpre_create). The payload's `user` includes `standard_attributes` and `custom_attributes`. | The project uses JWT access tokens.                      |
| A backend server, on demand                   | Admin API `User.standardAttributes` and `User.customAttributes`.                                                                                                              | None. The Admin API sees every attribute.                |
| A backend server, on change                   | Non-blocking event [user.profile.updated](../event.md#userprofileupdated), which fires after every `fill_form` submission.                                                    | A webhook or Deno hook subscribes to it.                 |

The ID token contains no attributes; see [ID Token](./design.md#id-token). The [resolver endpoint](../api-resolver.md) headers contain no attributes.

A backend must not assume a user has an attribute because a flow requires it. See [Meaning of required](#meaning-of-required).

## Auth UI

- Each form is one page. The page shows the form's [title and description](#translations), then its fields in order.
- Each field's input follows its `type`; see [user_profile field](#user_profile-field) and [fill_form_data](../authentication-flow-api-reference.md#fill_form_data). Labels use the attribute's existing [translations](#translations).
- Required fields are marked. The page cannot be submitted until they are filled.
- A page whose fields are all optional can be submitted with every field empty.
- Constraint errors and the hook's rejected attributes are shown under their field. The hook's `title` and `reason`, when present, are shown above the form.

## Backward compatibility

- Without `user_profile.forms`, the default flows are unchanged.
- `user.profile.pre_update` and `user.profile.updated` now also fire when a form is submitted. A hook that rejects every update with `triggered_by: user` now also blocks the form.

## Future works

- A `text` field kind, to show localized text between the fields of a form. The form's [description](#translations) covers text above all fields; `text` is for text next to one field. For example, a note before an optional consent:

  ```yaml
  user_profile:
    forms:
      signup:
        - fields:
            - user_profile:
                pointer: /x_accepted_terms_2026_01
                required: true
            - text:
                translation_key: marketing-note # "You can unsubscribe at any time in Settings."
            - user_profile:
                pointer: /x_marketing_consent
                required: false
  ```

  A `text` field takes no input and is never incomplete, so it never makes the form shown on its own.

- A consent step that records the accepted version and the time of acceptance, so a backend does not need to keep them from `user.profile.updated`. See [UC8](#uc8-agree-to-terms-and-conditions-at-signup).
- An embedded HTML step with proceed and discard actions.
- A step that redirects to an externally hosted page and resumes the flow.
- A flow builder in the portal.

## Appendix: competitor review

|                             | Auth0                                 | Entra External ID / Azure AD B2C                                         | Keycloak                            | Zitadel                              | Stytch                     |
| --------------------------- | ------------------------------------- | ------------------------------------------------------------------------ | ----------------------------------- | ------------------------------------ | -------------------------- |
| Mechanism                   | Forms shown by a post-login Action    | Attribute page in user flows; self-asserted pages in B2C custom policies | One realm-wide user profile         | Fixed register form; custom login UI | Build your own after login |
| Where "required" lives      | Action code                           | User flow; custom policy step                                            | The attribute, optionally per scope | Fixed fields                         | Nowhere                    |
| Prompt existing users       | Action code                           | B2C custom policies only                                                 | `VERIFY_PROFILE`, at every login    | No                                   | No                         |
| Multiple pages              | Yes                                   | B2C custom policies only                                                 | No                                  | No                                   | No                         |
| Validation hook             | Flow with a "Show error message" step | Submit hook with per-field errors                                        | Java validator                      | Actions v2 `interruptOnError`        | No                         |
| Same rules in the admin API | No                                    | No                                                                       | Yes                                 | Per API method                       | No                         |

References:

- https://auth0.com/docs/customize/forms/configure-additional-signup-steps
- https://learn.microsoft.com/en-us/entra/identity-platform/custom-extension-onattributecollectionsubmit-reference
- https://github.com/keycloak/keycloak/blob/main/docs/documentation/server_admin/topics/users/user-profile.adoc
- https://zitadel.com/docs/guides/integrate/onboarding/end-users
- https://stytch.com/docs/sdks/ui-configuration
