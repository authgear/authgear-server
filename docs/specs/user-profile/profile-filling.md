- [Profile Filling](#profile-filling)
  - [Goals](#goals)
  - [Non-goals](#non-goals)
  - [Configuration reference](#configuration-reference)
    - [Form object](#form-object)
    - [New authentication flow step: fill_form](#new-authentication-flow-step-fill_form)
    - [New identify step key: form](#new-identify-step-key-form)
    - [Translations](#translations)
  - [Use cases](#use-cases)
    - [UC1. Collect names at signup](#uc1-collect-names-at-signup)
    - [UC2. Collect names together with the username](#uc2-collect-names-together-with-the-username)
    - [UC3. Multi-page onboarding](#uc3-multi-page-onboarding)
    - [UC4. Validate input against an external system](#uc4-validate-input-against-an-external-system)
    - [UC5. Different applications need different attributes](#uc5-different-applications-need-different-attributes)
    - [UC6. Confirm attributes populated from an OAuth provider](#uc6-confirm-attributes-populated-from-an-oauth-provider)
    - [UC7. Collect attributes when an anonymous user is promoted](#uc7-collect-attributes-when-an-anonymous-user-is-promoted)
    - [UC8. Agree to terms and conditions at signup](#uc8-agree-to-terms-and-conditions-at-signup)
    - [UC9. Customizing the additional form order](#uc9-customizing-the-additional-form-order)
  - [The fill_form step](#the-fill_form-step)
    - [user_profile field](#user_profile-field)
      - [Fillable attributes](#fillable-attributes)
    - [Meaning of required](#meaning-of-required)
    - [When the form is shown](#when-the-form-is-shown)
    - [Pagination](#pagination)
    - [Config validation](#config-validation)
  - [Forms in the identify step](#forms-in-the-identify-step)
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

Profile Filling lets the developer collect [Standard Attributes](../glossary.md#standard-attributes) and [Custom Attributes](../glossary.md#custom-attributes) with [forms](../glossary.md#form) in a custom [Authentication Flow](../glossary.md#authentication-flow). A form is a list of fields. A [`fill_form` step](#the-fill_form-step) shows a form on its own page; an [`identify` step](#forms-in-the-identify-step) can show a form on the same page as the login ID.

Forms are available only in flows defined in `authentication_flow`. The flows Authgear generates from the project config have no forms. To add forms to them, define a flow named `default`, which replaces the generated flow of its type.

Whether an attribute is required is a property of a form, not of the attribute. A required attribute must be filled to complete that flow. It does not guarantee that every user has the attribute: users who signed up before the form existed, users created by the Admin API, and users who cleared the attribute in the settings page may not have it.

## Goals

- Collect attributes in custom signup, promote, and login flows, in Auth UI and in Custom UI.
- Collect attributes on the same page as the login ID at signup.
- Split the collected attributes across multiple pages.
- Apply the same validation in the flow, in the settings page, and in the Admin API.
- Let a hook reject input with per-attribute, localized messages.

## Non-goals

- Forms in the flows Authgear generates. See [Future works](#future-works).
- Collecting attributes in reauth flows and account recovery flows.
- Collecting [identities](../glossary.md#identity). See [Fillable attributes](#fillable-attributes).
- Making an attribute mandatory for every user.
- Showing a form only to users who lack an attribute. See [Future works](#future-works).
- Step types other than attribute forms. See [Future works](#future-works).

## Configuration reference

The configuration this spec adds, and where its rules are. Attribute definitions, including the `boolean` type and string constraints, are in [the user profile spec](./design.md).

### Form object

A Form is a list of fields. It is the `form` of a [fill_form step](#new-authentication-flow-step-fill_form) or of an [identify step branch](#new-identify-step-key-form).

| Key                              | Type                                                      | Default | Meaning                                                                                                                                                         | Rules                                       |
| -------------------------------- | --------------------------------------------------------- | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| `name`                           | string matching `^[a-zA-Z_][a-zA-Z0-9_]*$`, not `default` | none    | Names the form, so it can have its own [title and description](#translations), and so a [hook](#validation-hook) can tell forms apart. Names need not be unique. | [Translations](#translations)               |
| `fields`                         | list, non-empty                                           | —       | The fields of the form, in display order. Each item has exactly one key, which names the kind of field.                                                         | [The fill_form step](#the-fill_form-step)   |
| `fields[].user_profile`          | object                                                    | —       | A field that reads and writes one attribute.                                                                                                                    | [user_profile field](#user_profile-field)   |
| `fields[].user_profile.pointer`  | JSON pointer                                              | —       | The attribute. It must be [fillable](#fillable-attributes).                                                                                                     | [Fillable attributes](#fillable-attributes) |
| `fields[].user_profile.required` | boolean                                                   | `false` | Whether the attribute must be filled to submit the form.                                                                                                        | [Meaning of required](#meaning-of-required) |

### New authentication flow step: fill_form

A step type that shows one [Form](#form-object) on its own page, at its position in the flow.

It is available in:

- `authentication_flow.signup_flows[].steps[]`
- `authentication_flow.promote_flows[].steps[]`
- `authentication_flow.login_flows[].steps[]`
- the nested `steps` of the branches of those flows

| Key    | Type                        | Default | Meaning                                                                 |
| ------ | --------------------------- | ------- | ----------------------------------------------------------------------- |
| `type` | `fill_form`                 | —       | Required. Marks the step as a form.                                     |
| `name` | string                      | none    | The step's name, as for any step. It is unrelated to the form's `name`. |
| `form` | [Form object](#form-object) | —       | Required. The form to show.                                             |

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
            fields:
              - user_profile:
                  pointer: /x_company
                  required: true
```

Rules: [The fill_form step](#the-fill_form-step) and [Config validation](#config-validation).

### New identify step key: form

A key of the branches of an `identify` step. It shows a [Form](#form-object) on the same page as the login ID.

It is available in:

- `authentication_flow.signup_flows[].steps[].one_of[]`, where `type` is `identify`
- `authentication_flow.promote_flows[].steps[].one_of[]`, where `type` is `identify`
- the same steps nested in the branches of those flows

| Key    | Type                        | Default | Meaning                                                                                      |
| ------ | --------------------------- | ------- | -------------------------------------------------------------------------------------------- |
| `form` | [Form object](#form-object) | none    | A form filled together with the login ID. Allowed only when `identification` is `email`, `phone`, or `username`. |

```yaml
authentication_flow:
  signup_flows:
    - name: default
      steps:
        - type: identify
          one_of:
            - identification: username
              form:
                fields:
                  - user_profile:
                      pointer: /given_name
                      required: true
```

Rules: [Forms in the identify step](#forms-in-the-identify-step) and [Config validation](#config-validation).

### Translations

Each form page shows a title and a description above its fields:

| Key                                     | Meaning                                                                                                                                                                                                                                                            |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `v2.page.fill-form.default.title`       | The title of every form. Authgear provides it, for example "Complete your profile".                                                                                                                                                                                |
| `v2.page.fill-form.default.description` | The description of every form. Authgear provides it, worded by flow type, for example "Please provide the following information to complete your signup." in signup and promote flows, and "Please provide the following information to continue." in login flows. |
| `v2.page.fill-form.{name}.title`        | The title of the forms with that `name`. Falls back to `v2.page.fill-form.default.title`.                                                                                                                                                                          |
| `v2.page.fill-form.{name}.description`  | The description of the forms with that `name`. Falls back to `v2.page.fill-form.default.description`.                                                                                                                                                              |

The developer can override the built-in keys to change the text of every form, or set the `{name}` keys for one form. An empty translation hides the description.

A form in an `identify` step has no page of its own, so its title and description are not shown; the page keeps the identify page's title.

These keys are used by Auth UI. The Authflow API returns the form's `name`, as `form.name`, instead of the text, so a Custom UI shows its own title and description; see [fill_form_data](../authentication-flow-api-reference.md#fill_form_data).

Fields use the existing translation keys of their attributes. Labels and `enum` option labels are defined in [Defining custom attributes](./design.md#defining-custom-attributes) and [Custom Attribute type enum](./design.md#custom-attribute-type-enum). The message for a failed `pattern` is defined in [String constraints](./design.md#string-constraints).

The hook response and the Authflow API are not configuration; see [Validation hook](#validation-hook) and [the API reference](../authentication-flow-api-reference.md#type-signup-actiontype-fill_form).

## Use cases

Each use case shows the configuration, what the end-user sees, what a Custom UI exchanges with the Authflow API, and where the developer finds the result.

Every use case defines a flow named `default`, so it replaces the generated flow of that type. Flows of other types stay generated.

### UC1. Collect names at signup

A shopping app signs end-users up with email and password. It wants to greet end-users by name.

**Configuration**

```yaml
authentication_flow:
  signup_flows:
    - name: default
      steps:
        - name: setup_email
          type: identify
          one_of:
            - identification: email
        - type: verify
          target_step: setup_email
        - type: create_authenticator
          one_of:
            - authentication: primary_password
        - type: fill_form
          form:
            fields:
              - user_profile:
                  pointer: /given_name
                  required: true
              - user_profile:
                  pointer: /family_name
                  required: true
```

**What the end-user sees**

1. They enter an email address, verify it, and set a password.
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
        "form": {
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
- Webhook `user.created`, with `payload.user.standard_attributes.given_name` set to `"John"`.

### UC2. Collect names together with the username

A community app signs end-users up with a username and a password. It wants the end-user's name on the same page as the username, so signup has no extra page.

**Configuration**

```yaml
authentication_flow:
  signup_flows:
    - name: default
      steps:
        - type: identify
          one_of:
            - identification: username
              form:
                fields:
                  - user_profile:
                      pointer: /given_name
                      required: true
                  - user_profile:
                      pointer: /family_name
                      required: false
        - type: create_authenticator
          one_of:
            - authentication: primary_password
```

**What the end-user sees**

1. The first page shows Username, Given name, and Family name. Username and Given name are marked required.
2. They enter `johndoe` and "John", leave Family name empty, and continue.
3. They set a password and are signed in.

If `johndoe` is taken, the page shows the error under Username and keeps what they entered in the other fields.

**Authflow API**

The `identify` action carries the form in the option:

```json
{
  "type": "identification_data",
  "options": [
    {
      "identification": "username",
      "form": {
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
            "required": false
          }
        ]
      }
    }
  ]
}
```

The input adds `fields` to the usual identify input:

```json
{
  "identification": "username",
  "login_id": "johndoe",
  "fields": [{ "key": "/given_name", "value": "John" }]
}
```

The next action is `create_authenticator`. Sending the input without `/given_name` fails with `ValidationFailed`, and the flow stays in the `identify` step.

**Where the attribute appears**

The same places as in [UC1](#uc1-collect-names-at-signup).

### UC3. Multi-page onboarding

A fitness app asks for personal details first and preferences second. Preferences are optional.

**Configuration** (`/x_fitness_goal` and `/x_referral_source` are `enum` custom attributes with `end_user: readwrite`)

```yaml
authentication_flow:
  signup_flows:
    - name: default
      steps:
        # identify, verify, and create_authenticator as in UC1
        - type: fill_form
          form:
            fields:
              - user_profile:
                  pointer: /given_name
                  required: true
              - user_profile:
                  pointer: /birthdate
                  required: true
        - type: fill_form
          form:
            fields:
              - user_profile:
                  pointer: /x_fitness_goal
                  required: false
              - user_profile:
                  pointer: /x_referral_source
                  required: false
```

Each `fill_form` step is one page.

**What the end-user sees**

1. Page 1: Given name and a date picker for Birthdate, both required.
2. Page 2: two dropdowns, Fitness goal and How did you hear about us, both optional. They can continue with both empty.

**Authflow API**

Two consecutive `fill_form` actions, each with its own `state_token`. Submitting page 2 with `{"fields": []}` is valid, because nothing on it is required.

**Where the attribute appears**

- User Info endpoint: `"birthdate": "1990-05-01"` at the root. If the end-user filled page 2, `custom_attributes` also has `x_fitness_goal`.
- An optional attribute left empty is absent from the response, not `null`.

### UC4. Validate input against an external system

An employee portal collects an employee ID. The format is checked by an [attribute constraint](#attribute-constraints). Whether the ID exists is checked by a [validation hook](#validation-hook) that calls the HR system. The hook reads `payload.user.custom_attributes.x_employee_id`, which both events carry.

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
authentication_flow:
  signup_flows:
    - name: default
      steps:
        # identify, verify, and create_authenticator as in UC1
        - type: fill_form
          form:
            name: employee
            fields:
              - user_profile:
                  pointer: /x_employee_id
                  required: true
hook:
  blocking_handlers:
    - event: authentication.form.post_submitted
      url: https://hr-check.example.com/hook
    - event: user.profile.pre_update
      url: https://hr-check.example.com/hook
```

The translation `custom-attribute-pattern-error-/x_employee_id` is set to "Format: E followed by 5 digits".

The hook receives `authentication.form.post_submitted` with `payload.form.name` set to `employee`. A hook that serves several forms uses it to decide which checks to run.

**What the end-user sees**

- They type `12345`. The page shows "Format: E followed by 5 digits" under the field.
- They type `E99999`, which has the right format but doesn't exist. The hook answers in `context.language`, here `zh-HK`:

  ```json
  {
    "is_allowed": false,
    "title": "找不到員工",
    "reasons": [
      {
        "type": "invalid_form_field",
        "field": {
          "user_profile": {
            "pointer": "/x_employee_id"
          }
        },
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
- If an admin later sets `E99999` through the Admin API, the hook receives `user.profile.pre_update` with `context.triggered_by: admin_api`, and the mutation fails with `HookDisallowed`.

### UC5. Different applications need different attributes

One project serves a store and a forum. The store needs a shipping address; the forum needs a nickname.

Each application gets its own signup flow, selected by a flow group.

**Configuration**

```yaml
authentication_flow:
  signup_flows:
    - name: store
      steps:
        # identify, verify, and create_authenticator as in UC1
        - type: fill_form
          form:
            fields:
              - user_profile:
                  pointer: /address
                  required: true
    - name: forum
      steps:
        # identify, verify, and create_authenticator as in UC1
        - type: fill_form
          form:
            fields:
              - user_profile:
                  pointer: /nickname
                  required: true
ui:
  authentication_flow:
    groups:
      - name: store
        flows:
          - type: signup
            name: store
          - type: login
            name: default
      - name: forum
        flows:
          - type: signup
            name: forum
          - type: login
            name: default
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
2. They sign up in the store with another account and are asked only for an address: street, city, region, postal code, country.

A user who signed up in the forum and later signs in to the store is not asked for an address. Asking only users who lack it needs [`show`](#future-works).

**Authflow API**

A Custom UI for the store creates the flow with `{"type": "signup", "name": "store"}`. The `fill_form` action lists `/address` once. The input sends it as one object:

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

- User Info endpoint: `"nickname"` or `"address": { ... }` at the root. Both the store and the forum see every attribute the user has, because attributes belong to the user, not to the application.

### UC6. Confirm attributes populated from an OAuth provider

A travel app lets end-users sign up with Google. It wants a correct legal name, so the end-user reviews the name Google returned before the account is created.

**How it works**

1. The `identify` step creates the Google identity. Because of [population](./design.md#standard-attributes-population) with `strategy: on_signup`, that step also copies Google's `given_name` and `family_name` claims into the new user's standard attributes.
2. The next step is a `fill_form` step. Each field shows the attribute's current value, so the page opens with the Google values already filled in.
3. The end-user keeps or edits the values and submits. The submitted values replace the populated ones.

If the end-user signs up with email instead, population has nothing to copy, and the same page opens with empty fields.

**Configuration**

```yaml
user_profile:
  standard_attributes:
    population:
      strategy: on_signup
authentication_flow:
  signup_flows:
    - name: default
      steps:
        - name: identify
          type: identify
          one_of:
            - identification: oauth
            - identification: email
              steps:
                - type: verify
                  target_step: identify
                - type: create_authenticator
                  one_of:
                    - authentication: primary_password
        - type: fill_form
          form:
            fields:
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

Google is set up as an OAuth provider in `identity.oauth.providers`.

**What the end-user sees**

1. They select "Continue with Google" and approve on Google's page.
2. Back in Auth UI, a page shows Given name "John" and Family name "Doe", taken from Google. Gender is empty.
3. They change the family name to "Doe-Smith" and continue. They are redirected to the app, signed in.

**Authflow API**

After the OAuth callback, `data.form.fields` carries the populated values:

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
authentication_flow:
  promote_flows:
    - name: default
      steps:
        # identify, verify, and create_authenticator as in UC1
        - type: fill_form
          form:
            fields:
              - user_profile:
                  pointer: /nickname
                  required: true
```

Only the promote flow has the form, so players who sign up directly are not asked.

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
authentication_flow:
  signup_flows:
    - name: default
      steps:
        # identify, verify, and create_authenticator as in UC1
        - type: fill_form
          form:
            fields:
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

The pointer includes the terms version, so a later version gets a new attribute.

**What the end-user sees**

1. They enter an email address and a password.
2. A page shows two checkboxes. The first reads "I agree to the Terms of Service and Privacy Policy"; both names are links that open in a new tab. The first checkbox is marked required.
3. With the first checkbox unchecked, the page cannot be submitted.
4. They check the first box, leave the marketing box unchecked, and continue.

**Authflow API**

`data.form.fields`:

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
- Webhook `user.created`. A backend that must keep proof of consent stores `context.timestamp` and the attribute value from this event, because Authgear keeps the current value only.
- Settings page: the end-user can uncheck either box. Unchecking the terms box does not delete the account.

### UC9. Customizing the additional form order

An employee app signs users up with a phone number and an SMS OTP. Each SMS costs money, so the app checks the employee ID before sending one. The form goes right after the phone number is entered, before the OTP.

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

To ask for the employee ID on the same page as the phone number, put the form in the `identify` step instead, as in [UC2](#uc2-collect-names-together-with-the-username).

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

A `fill_form` step shows one form on its own page. Where it is available is listed in the [configuration reference](#new-authentication-flow-step-fill_form).

```yaml
- type: fill_form
  form:
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

The rules in this section apply to every form, including a form in an [`identify` step](#forms-in-the-identify-step), except where that section says otherwise.

### user_profile field

A `user_profile` field reads and writes one attribute. It follows the [User Profile Pointer](../convention.md#user-profile-pointer) convention.

Its keys, `pointer` and `required`, are listed in the [configuration reference](#form-object).

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

### Meaning of required

A required `user_profile` field means the form cannot be submitted until the attribute is filled. For a `boolean` attribute, the checkbox must be checked. It has no effect outside the flow:

- Users who never ran the flow may lack the attribute.
- The Admin API can create users without it or clear it.
- The end-user can clear it in the settings page.

### When the form is shown

A form is shown every time the flow reaches it, in every flow type. A form in a login flow is shown at every login.

The form includes every field, optional and already-filled ones included. Filled fields show their current value, including values [populated](./design.md#standard-attributes-population) by an earlier `identify` step; see [UC6](#uc6-confirm-attributes-populated-from-an-oauth-provider). A current value is checked against the current constraints on submission like any other, so a value that no longer meets them must be corrected, or cleared if the field is optional, before the form can be submitted.

When a signup flow continues as an existing user through [account linking](../account-linking.md), its forms are skipped, because the flow no longer sets up a new user.

### Pagination

Each `fill_form` step is one page. To split the fields across pages, write consecutive `fill_form` steps.

### Config validation

A config is rejected when a form:

- has a `user_profile` field whose attribute is not [fillable](#fillable-attributes);
- has two `user_profile` fields with the same pointer; or
- is named `default`, because `v2.page.fill-form.default.*` are the built-in keys.

A config is also rejected when:

- a `fill_form` step can be reached, on any path through its flow, before the end-user is identified in a signup or promote flow, or before the end-user is authenticated in a login flow. Otherwise the form could write to the profile of a user who has not proven who they are, and the [validation hook](#validation-hook) would receive a user without identities.
- an `identify` branch has a `form` and its `identification` is not `email`, `phone`, or `username`, or its flow is not a signup or promote flow. See [Forms in the identify step](#forms-in-the-identify-step).

A change to access control that makes an attribute used by any form non-fillable is also rejected.

## Forms in the identify step

A branch of an `identify` step in a signup or promote flow can have a `form`. Its fields are filled on the same page as the login ID, and submitted in the same input. Its keys are in the [configuration reference](#new-identify-step-key-form).

| Rule                                                                                                                                               | Reason                                                                                                                                                                  |
| -------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Only `email`, `phone`, and `username` branches can have a `form`.                                                                                  | Other identifications leave the page, such as to an OAuth provider, so there is no page to put the fields on. Put a `fill_form` step after the `identify` step instead. |
| Not available in login, signup_login, reauth, or account recovery flows.                                                                           | Before identification and authentication, the end-user has not proven who they are.                                                                                     |
| Each branch has its own form. A branch without `form` collects no attributes.                                                                       | Branches are different signup methods, such as email and phone, and may need different fields.                                                                          |
| The login ID is processed first, as it is without a form, including [authentication.post_identified](../event.md#authenticationpost_identified). Then the form is processed as a submitted form. | An unavailable login ID is reported without calling the form's hook.                                                                                                    |
| If the login ID or the form is rejected, the flow stays in the `identify` step, and the end-user submits both again.                              | The two are one input.                                                                                                                                                  |
| When a signup_login flow continues into the signup flow, the login ID is already entered, so the form is shown on its own page right after it, as a `fill_form` step. | The signup_login page cannot know whether the end-user signs up or logs in before the login ID is entered.                                                             |

Otherwise the form follows the rules of [the fill_form step](#the-fill_form-step). In the Authflow API, the option carries the form, and the input carries its fields; see [Form in identification options](../authentication-flow-api-reference.md#form-in-identification-options). See [UC2](#uc2-collect-names-together-with-the-username) for an example.

## Where profile filling happens

Profile filling happens only inside a flow that has a form:

| Entry point                                                                         | Profile filling                                                                                 |
| ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Signup, promote, and login flows, in Auth UI and Custom UI                          | The flow's forms.                                                                               |
| Continuing an existing session in Auth UI                                           | The forms of the login flow Auth UI would use for the authorization request.                    |
| [`select_account`](../custom-ui-select-account.md) in Custom UI                     | The login flow's later steps, including its forms.                                              |
| Reauth flows, account recovery flows                                                | None.                                                                                           |
| `prompt=none`, refresh token grant, biometric login, app2app, pre-authenticated URL | None. These do not run a login flow.                                                            |

## Validation

### Attribute constraints

Values are checked against the attribute's own constraints, on every write path:

- `minimum` and `maximum` on [`integer`](./design.md#custom-attribute-type-integer) and [`number`](./design.md#custom-attribute-type-number) custom attributes;
- [string constraints](./design.md#string-constraints): `min_length`, `max_length`, `allowed_characters`, and `pattern`;
- the fixed formats of standard attributes and other custom attribute types.

Auth UI shows a failed constraint's message under the field.

### Validation hook

A flow stores the filled attributes only when it finishes, so two blocking events apply:

| When                                        | Event                                                                                                                                                  |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| A form is submitted                         | [authentication.form.post_submitted](../event.md#authenticationformpost_submitted)                                                                     |
| The flow finishes and stores the attributes | [user.pre_create](../event.md#userpre_create) in a signup flow; [user.profile.pre_update](../event.md#userprofilepre_update) in a login or promote flow |

Both payloads contain the user with the submitted attributes applied. Only a rejection on form submission can be shown under the fields, so per-attribute checks belong there. A rejection when the flow finishes fails the last input of the flow, and the end-user cannot correct the form.

To reject specific fields of a submitted form, the hook returns [`reasons`](../event.md#authenticationformpost_submitted) of type `invalid_form_field`. See [UC4](#uc4-validate-input-against-an-external-system) for an example. In the Authflow API, the `HookDisallowed` error carries them in the `reasons` of the hook's entry in `info.reasons`, each as `{ "type": "invalid_form_field", "key", "message" }`, where `key` is the key of the matching field.

The hook localizes its `title`, `reason`, and messages using `context.language`; Authgear displays them unchanged.

### Where validation applies

| Write path                                    | Attribute constraints | Validation hook                                                                                                                                                   | `required`   |
| --------------------------------------------- | --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ |
| A form in any flow                            | ✅                    | `authentication.form.post_submitted` when the form is submitted; `user.pre_create` or `user.profile.pre_update` when the flow finishes. See [Validation hook](#validation-hook). | Per form     |
| Settings page                                 | ✅                    | `user.profile.pre_update`                                                                                                                                         | Not enforced |
| Admin API and portal                          | ✅                    | `user.profile.pre_update`, with `context.triggered_by` set to `admin_api` or `portal`                                                                             | Not enforced |

## Reading the filled attributes

Filled attributes are stored like any other attribute write. Other components read them through the existing channels:

| Reader                                        | Channel                                                                                                                                                                                                         | Condition                                                |
| --------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------- |
| The application, with the user's access token | [User Info endpoint](./design.md#user-info-endpoint). Standard attributes are at the root; custom attributes are under `custom_attributes`.                                                                     | The attribute's `bearer` access control is not `hidden`. |
| A backend server, per request                 | JWT access token claims added by a hook on [oidc.jwt.pre_create](../event.md#oidcjwtpre_create). The payload's `user` includes `standard_attributes` and `custom_attributes`.                                   | The project uses JWT access tokens.                      |
| A backend server, on demand                   | Admin API `User.standardAttributes` and `User.customAttributes`.                                                                                                                                                | None. The Admin API sees every attribute.                |
| A backend server, on change                   | Non-blocking event [user.created](../event.md#usercreated) after a signup flow, or [user.profile.updated](../event.md#userprofileupdated) after a login or promote flow with a submitted form.                   | A webhook or Deno hook subscribes to it.                 |

The ID token contains no attributes; see [ID Token](./design.md#id-token). The [resolver endpoint](../api-resolver.md) headers contain no attributes.

A backend must not assume a user has an attribute because a flow requires it. See [Meaning of required](#meaning-of-required).

## Auth UI

- Each `fill_form` step is one page. The page shows the form's [title and description](#translations), then its fields in order.
- A form in an `identify` step is shown below the login ID input, without its title and description. When the step has several branches with forms, each branch's input shows its own fields.
- Each field's input follows its `type`; see [user_profile field](#user_profile-field) and [fill_form_data](../authentication-flow-api-reference.md#fill_form_data). Labels use the attribute's existing [translations](#translations).
- Required fields are marked. The page cannot be submitted until they are filled.
- A page whose fields are all optional can be submitted with every field empty.
- Constraint errors and the hook's rejected attributes are shown under their field. The hook's `title` and `reason`, when present, are shown above the form.

## Backward compatibility

- Flows without forms are unchanged.
- `user.profile.pre_update` and `user.profile.updated` now also fire when a login or promote flow in which the end-user submitted a form finishes. A hook that rejects every update with `triggered_by: user` now also blocks those flows.

## Future works

- Forms in the flows Authgear generates, so a project can collect attributes without defining its own flows.
- A form key `show` with the values `always` and `when_incomplete`. With `when_incomplete`, the form is shown only when one of its required fields is not filled. A login flow could then ask existing users for a newly needed attribute, or to accept updated terms, once instead of at every login, and an application could ask for the attributes it needs from users who signed up elsewhere.
- A `text` field kind, to show localized text between the fields of a form. The form's [description](#translations) covers text above all fields; `text` is for text next to one field. For example, a note before an optional consent:

  ```yaml
  - type: fill_form
    form:
      fields:
        - user_profile:
            pointer: /x_accepted_terms_2026_01
            required: true
        - text:
            translation_key: marketing-note # "You can unsubscribe at any time in Settings."
        - user_profile:
            pointer: /x_marketing_consent
            required: false
  ```

- A consent step that records the accepted version and the time of acceptance, so a backend does not need to keep them from `user.created`. See [UC8](#uc8-agree-to-terms-and-conditions-at-signup).
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
