# End-to-End tests

This directory contains all resources related to end-to-end tests. It starts a docker-compose environment with a running authgear server at port `localhost:4000` and runs tests against it.

The end-to-end tests are run in CI/CD `authgear-e2e` job.

## Running the tests

Simply run the following command:

```bash
make run
```

## Authflow tests

The authflow tests are located in the `tests` directory. Unlike normal golang tests, these tests are written in YAML.

An example test case is as follows:

```yaml
# login/test.yaml

# Name of the test case
name: Email Login

authgear.yaml:
  # Which authgear.yaml to extend on (Optional)
  extend: base.yaml

  # Override specific fields in the authgear.yaml
  # (Merges on maps and replaces on arrays/scalars)
  override: |
    authentication:
        identities:
          - login_id
        primary_authenticators:
          - password

# Before steps to run before the test, used for creating fixtures
before:
  # Uses user import API (/docs/specs/user-import.md) to import users from a JSON file
  - type: user_import
    user_import: users.json

  # Executes a custom SQL script, useful for data that cannot be achieved through the user import API
  # e.g. Password expiry
  - type: customsql
    customsql:
      file: ./custom.sql

# Test steps, each step is an API call to Authflow API (/docs/specs/authentication-flow-api-reference.md)
steps:
  # Use `action: create` to create a flow
  - name: Start login
    action: "create"
    input: |
      {
        "type": "login",
        "name": "default"
      }
    # `result` is the expected response from the API
    output:
      # [[string]] is a specific matcher pattern that matches any string, useful for dynamic values e.g. OTP code
      result: |
        {
          "state_token": "[[string]]",
          "type": "login",
          "name": "default",
          "action": {
              "type": "identify",
              "data": {
                  "type": "identification_data",
                  "options": "[[array]]"
              }
          }
        }

  # Use `action: input` to input data and proceed to the next step
  - name: Choose to login with username
    action: input
    input: |
      {
        "identification": "username",
        "login_id": "e2e_login"
      }
    output:
      result: |
        {
          "action": {
            "type": "authenticate"
          }
        }

  # You can also use `error` to check for expected errors
  - name: Enter incorrect password
    action: input
    input: |
      {
        "authentication": "primary_password",
        "password": "incorrect_password"
      }
    output:
      # `error` is the expected error response from the API
      error: |
        {
          "reason": "InvalidCredentials"
        }

  # Finally, use `type: finish` to finish the flow
  - name: Enter correct password
    action: input
    input: |
      {
        "authentication": "primary_password",
        "password": "correct_password"
      }
    output:
      result: |
        {
          "action": {
            "type": "finished"
          }
        }
```

### Before steps

Before steps are used to create fixtures for the test. The following before steps are available:

- `type: user_import`: Uses user import API to import users from a JSON file
- `type: customsql`: Executes a custom SQL script

#### User import

```yaml
- type: user_import
  user_import: users.json
```

The format of `users.json` is documented in [user-import.md](/docs/specs/user-import.md#the-input-format)

#### Custom SQL

```yaml
- type: customsql
  customsql:
    file: ./custom.sql
```

`custom.sql` is preprocessed as a Go template with the following variables:

- `{{ .AppID }}`: The database URL

### Steps

Each step is an API call to the Authflow API. The following steps are available:

- `action: create`: Creates a flow
- `action: input`: Inputs data and proceeds to the next step
- `action: oauth_redirect`: Redirects to an OAuth provider
- `action: query`: Query from database

#### OAuth redirect

The `oauth_redirect` action is used to redirect to an OAuth provider. The result is `prev.result.code` that can be used in the next step to finish the identification.

```yaml
- name: Redirect to Google
  action: oauth_redirect
  input: |
    {
      "provider_id": "google"
    }

- action: input
  input: |
    {
      "code": "{{ .prev.result.code }}"
    }
  output:
    result: |
      {
        "action": {
          "type": "finished"
        }
      }
```

#### Database Query

The `query` action is used to query from database. The results can be accessed by `prev` in next step.

```yaml
- action: query
  query: |
    SELECT id
    FROM _auth_user 
    WHERE app_id = '{{ .AppID }}'
    AND standard_attributes ->> 'preferred_username' = 'my_username';
  query_output:
    rows: |
      [
        {
          "id": "[[string]]"
        }
      ]
- action: input
  input: |
    {
      "identification": "id_token",
      "id_token": "{{ generateIDToken (index .prev.result.rows 0).id }}"
    }
```

#### Assert

The `output` field is used to assert the response or error from the API.

```yaml
output:
  result: |
    {
      "state_token": "[[string]]",
      "type": "login",
      "name": "default",
      "action": {
          "type": "identify",
          "data": {
              "type": "identification_data",
              "options": "[[array]]"
          }
      }
    }
  error: |
    {
      "reason": "InvalidCredentials"
    }
```

The following matchers are available for assertions:

- `[[string]]`: Matches any string
- `[[array]]`: Matches any array
- `[[object]]`: Matches any object
- `[[number]]`: Matches any number
- `[[boolean]]`: Matches any boolean
- `[[null]]`: Matches null
- `[[never]]`: Disallows the field, useful for blacklisting fields in maps
- `["[[arrayof]]", "[[object]]"]`: Matches an array of objects, 2nd element can be any matcher or a specific value
- `["[[string]]", "some_constant"]`: Matches a tuple. Extra or missing elements are reported as errors

### JSON schema

The test cases are validated against a JSON schema located at [schema.json](./schema.json).

For example, in VSCode, you can use the [YAML extension](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml) to enable schema validation.

```json
{
  "yaml.schemas": {
    "./e2e/schema.json": "e2e/**/*test.yaml"
  }
}
```
