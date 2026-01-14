* [Contributing guide](#contributing-guide)
  * [Install dependencies](#install-dependencies)
    * [Install dependencies with asdf and homebrew](#install-dependencies-with-asdf-and-homebrew)
    * [Install dependencies with Nix Flakes](#install-dependencies-with-nix-flakes)
  * [Set up environment](#set-up-environment)
  * [Set up the database](#set-up-the-database)
  * [Set up MinIO](#set-up-minio)
  * [Run](#run)
  * [Create an account for yourselves and grant you access to the portal](#create-an-account-for-yourselves-and-grant-you-access-to-the-portal)
  * [Known issues](#known-issues)
    * [Known issues on portal](#known-issues-on-portal)
  * [Comment tags](#comment-tags)
  * [Common tasks](#common-tasks)
    * [How to create a new database migration?](#how-to-create-a-new-database-migration)
    * [Set up HTTPS to develop some specific features](#set-up-https-to-develop-some-specific-features)
    * [Create release tag for a deployment](#create-release-tag-for-a-deployment)
    * [Keep dependencies up\-to\-date](#keep-dependencies-up-to-date)
  * [Generate translation](#generate-translation)
  * [Set up LDAP for local development](#set-up-ldap-for-local-development)
    * [Create a LDAP user](#create-a-ldap-user)
    * [Configure Authgear](#configure-authgear)
    * [Start with the profile ldap](#start-with-the-profile-ldap)
  * [Switching between sessionType=refresh\_token and sessionType=cookie](#switching-between-sessiontyperefresh_token-and-sessiontypecookie)
  * [Switch to Database config source](#switch-to-database-config-source)
* [Storybooks](#storybooks)
* [Agentic coding](#agentic-coding)
  * [Always prompts](#always-prompts)
  * [Manual prompts](#manual-prompts)

# Contributing guide

This guide teaches you to run Authgear locally for development purpose.
It also covers some common development tasks.

## Install dependencies

### Install dependencies with asdf and homebrew

This project supports asdf, and there is a .tool-versions file at the project root.

1. Install asdf
2. Run the following to install all dependencies in .tool-versions

   ```sh
   asdf plugin add golang https://github.com/asdf-community/asdf-golang.git
   asdf plugin add nodejs https://github.com/asdf-vm/asdf-nodejs.git
   asdf plugin add python
   asdf install
   ```
3. Install pkg-config

   ```sh
   brew install pkg-config
   ```

4. Install icu4c

   ```sh
   brew install icu4c
   ```

   icu4c installed by brew is not globally visible by default, so you have to ensure your shell has the following in effect

   ```sh
   export PKG_CONFIG_PATH="$(brew --prefix)/opt/icu4c/lib/pkgconfig"
   ```

   To avoid doing the above every time you open a new shell, you may want to add it to your shell initialization script such as `~/.profile`, `~/.bash_profile`, etc.

5. Install libvips

   ```sh
   brew install vips
   ```

   libvips on macOS requires `-Xpreprocessor` to build.
   Run the following to tell Cgo.

   ```sh
   export CGO_CFLAGS_ALLOW="-Xpreprocessor"
   ```

6. Install libmagic

   ```sh
   brew install libmagic
   ```

   Run the following to tell Cgo where to find libmagic.
   Preferably you add it to your shell startup script.

   ```sh
   export CGO_CFLAGS="-I$(brew --prefix)/include"
   export CGO_LDFLAGS="-L$(brew --prefix)/lib"
   ```

7. Run `make vendor`.

### Install dependencies with Nix Flakes

This project supports Nix Flakes.
If you are not a Nix user, please see the above section instead.

1. Make a shell with dependencies installed.

You can either run

```sh
nix develop
```

Or if you are also a direnv and nix-direnv user, you can place a `.envrc` at the project
with the following content

```
# shellcheck shell=bash
if ! has nix_direnv_version || ! nix_direnv_version 3.0.6; then
  source_url "https://raw.githubusercontent.com/nix-community/nix-direnv/3.0.6/direnvrc" "sha256-RYcUJaRMf8oF5LznDrlCXbkOQrywm0HDv1VjYGaJGdM="
fi
use flake
```

2. Run `make build-frontend`.

## Set up environment

1. Set up environment variables

   ```sh
   cp .env.example .env
   ```

2. Generate config files

   ```sh
   $ go run ./cmd/authgear init \
      --interactive false \
      --output-folder ./var \
      --purpose portal \
      --app-id accounts \
      --public-origin 'http://accounts.portal.localhost:3100' \
      --portal-origin 'http://portal.localhost:8000' \
      --portal-client-id portal \
      --phone-otp-mode sms \
      --disable-email-verification true \
      --search-implementation postgresql
   ```

   You need to make changes according to the example shown above.

   `authgear.yaml` must contain the following contents for the portal to work.

   ```yaml
   oauth:
     clients:
     - client_id: portal
       redirect_uris:
       # See nginx.conf for the difference between 8000, 8001, 8010, and 8011
       - "http://portal.localhost:8000/oauth-redirect"
       - "http://portal.localhost:8001/oauth-redirect"
       - "http://portal.localhost:8010/oauth-redirect"
       - "http://portal.localhost:8011/oauth-redirect"
       # This redirect URI is used by the iOS and Android demo app.
       - "com.authgear.example://host/path"
       # This redirect URI is used by the React Native demo app.
       - "com.authgear.example.rn://host/path"
       # This redirect URI is used by the Flutter demo app.
       - com.authgear.exampleapp.flutter://host/path
       # This redirect URI is used by the Xamarin demo app.
       - com.authgear.exampleapp.xamarin://host/path
       post_logout_redirect_uris:
       # This redirect URI is used by the portal development server.
       - "http://portal.localhost:8000/"
       # This redirect URI is used by the portal production build.
       - "http://portal.localhost:8010/"
   ```

3. Set up `.localhost`

   For cookie to work properly, you need to use

   - `portal.localhost:8000` to access the portal.
   - `accounts.portal.localhost:3100` to access the main server.

   You can either do this by editing `/etc/hosts` or using `dnsmasq`.

## Set up the database

1. Start the database

   ```sh
   docker compose build postgres16
   docker compose up -d postgres16 pgbouncer
   ```

2. Apply migrations

   ```sh
   go run ./cmd/authgear database migrate up
   go run ./cmd/authgear audit database migrate up
   go run ./cmd/authgear images database migrate up
   go run ./cmd/authgear search database migrate up
   go run ./cmd/portal database migrate up
   ```


## Set up MinIO

```sh
docker compose up -d minio
docker compose exec -it minio bash

# Inside the container
mc alias set local http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"
# Create a bucket named "images"
mc mb local/images
# Create a bucket named "userexport"
mc mb local/userexport
```

## Run

1. In case you have made changes to authui, you run `make authui` to re-build the assets, or run `make authui-dev` to start the development server (Hot Reload/Hot Module Replacement supported).
2. In case you have not started the dependencies, run `docker compose up -d`.
3. Run `make start` to run the main server.
4. Run `make start-portal` to run the portal server.
5. `cd portal; npm start` to run the portal frontend.

## Create an account for yourselves and grant you access to the portal

1. Create an account

   The following command assumes everything is running, as it invokes the Admin API server to create an account!

   ```sh
   $ go run ./cmd/authgear internal admin-api invoke \
     --app-id accounts \
     --endpoint "http://localhost:3002" \
     --host "accounts.portal.localhost:3100" \
     --query '
       mutation createUser($email: String!, $password: String!) {
         createUser(input: {
           definition: {
             loginID: {
               key: "email"
               value: $email
             }
           }
           password: $password
         }) {
           user {
             id
           }
         }
       }
     ' \
     --variables-json "$(jq -cn --arg email "user@example.com" --arg password "password" '{email: $email, password: $password}')" | tee ./query_output
   ```

2. Make the account the owner of `accounts`

   ```sh
   decoded_node_id="$(jq <./query_output --raw-output '.data.createUser.user.id' | basenc --base64url --decode)"
   raw_id="${decoded_node_id#User:}"
   go run ./cmd/portal internal collaborator add \
      --app-id accounts \
      --user-id "$raw_id" \
      --role owner
   ```

3. Now you can navigate to your project in the [portal](http://portal.localhost:8000)

## Known issues

cert-manager@v1.7.3 has transitive dependency problem.

When intl-tel-input is >= 21, it switched to use CSS variables. https://github.com/jackocnr/intl-tel-input/releases/tag/v21.0.0
The problem is that it uses `--custom-var: url("../some-path");`, which is rejected by Parcel https://github.com/parcel-bundler/parcel/blob/v2.10.2/packages/transformers/css/src/CSSTransformer.js#L135

When intl-tel-input is >= 20, the behavior of initialCountry. It no longer supports selecting the first country in its own sorted list of countries when we pass `initialCountry=""`.

When intl-tel-input is >= 19, isPossibleNumber is removed. isValidNumber becomes isPossibleNumber. isValidNumberPrecise is the old isValidNumber.

So the highest version of intl-tel-input is 18.

### Known issues on portaligraphql

I tried to update to graphiql to v5 on 2025-11-21, but it has UI glitches and the CTRL+SPACE autocomplete no longer works.

On the other hand, the newly built bundle contains `{{` that confuses Go `html/template`.
This issue can be worked around by using simple string replacement, instead of `html/template`.
When we do that, we should change `{{ $.CSPNONCE }}` to something like `@@CSPNONCE@@` so that the reader will not be confused `html/template` is being used.

### Known issues on portal

NPM has an outstanding issue related to optional native dependencies.
https://github.com/npm/cli/issues/4828
The issue will happen if the following conditions hold:
- The package.json, package-lock.json and node\_modules are in correct state. node\_modules only contain macOS arm dependencies.
- We update the version of parcel and run npm install to update package-lock.json
- package-lock.json becomes invalid.
- npm ci becomes broken on non macOS arm machines
So whenever we want to update dependencies, we first delete node\_modules and package-lock.json.
Then npm install will generate a correct package-lock.json.

Docker Desktop on Mac has [an issue](https://github.com/docker/for-mac/issues/5812#issuecomment-874532024) that would lead to an unresponsive reverse proxy.
One of the comment says enabling "Use the new Virtualization framework" would help.
After >5000 requests to the portal, "upstream timed out" errors will begin to pop up.
If enabling "Use the new Virtualization framework" did not help, you can restart Docker Desktop on Mac as a workaround.

[Radix Icons] https://www.radix-ui.com/icons can only be imported in code, and no fonts is available. Currently, vite will select a most suitable chunk to bundle the icons.

## Comment tags

- `FIXME`: Should be fixed as soon as possible
- `TODO`: Should be done when someone really needs it.
- `OPTIMIZE`: Should be done when it really becomes a performance issue.
- `SECURITY`: Known potential security issue.

## Common tasks

### How to create a new database migration?

```sh
go run ./cmd/authgear database migrate new my_new_migration
```

### Set up HTTPS to develop some specific features

Usually you can use `localhost` to bypass the requirement of HTTPS.
In case you are developing a feature that does not allow you to do so,
You can use [mkcert](https://github.com/FiloSottile/mkcert).

```sh
# Install mkcert
brew install mkcert
# Install the root CA into Keychain Access.
mkcert -install
# Generate the TLS certificate
make mkcert
# Uncomment the TLS config in nginx.confg to enable TLS, restart nginx to apply the change.
```

Update `public_origin` in `var/authgear.yaml` to your local https domain.

### Create release tag for a deployment

```sh
# Create release tag when deploying to staging or production
# For staging, prefix the tag with `staging-`. e.g. staging-2021-05-06.0
# For production, no prefix is needed. e.g 2021-05-06.0
# If there are more than 1 release in the same day, increment the last number by 1
git tag -a YYYY-MM-DD.0

# Show the logs summary
make logs-summary A=<previous tag> B=<current tag>
```

### Keep dependencies up-to-date

Various files in this project have versioned dependencies.

- [The go directive in go.mod](./go.mod)
- [The dependencies listed in go.mod](./go.mod)
- [The tool versions listed in .tool-versions](./.tool-versions)
- [The version of golangci-lint in Makefile](./Makefile)
- [The versions appearing in ./github/workflows/ci.yaml](./github/workflows/ci.yaml)
- [The FROM directives in ./cmd/authgear/Dockerfile](./cmd/authgear/Dockerfile)
- [The FROM directives in ./cmd/portal/Dockerfile](./cmd/portal/Dockerfile)
- [The dependencies in ./authui/package.json](./authui/package.json)
- [The dependencies in ./portal/package.json](./portal/package.json)
- [The dependencies in ./scripts/npm/package.json](./scripts/npm/package.json)
  - Note that you cannot simply upgrade `tzdata` because the version must match that of the server.
  - The server version is documented and enforced in the Dockerfiles during the build process.
  - Visit [https://github.com/rogierschouten/tzdata-generate/releases](https://github.com/rogierschouten/tzdata-generate/releases) to see which `tzdata` version correspond to which server version.
- [The cropperjs type definition in ./authui/src](./authui/src)
- [./pkg/util/geoip/GeoLite2-Country.mmdb](./pkg/util/geoip/GeoLite2-Country.mmdb)
- [GraphiQL](./pkg/util/graphqlutil/graphiql.go)
- [Material Icons](authui/src/authflowv2/icons/material-symbols-outlined.woff2)
  - Download the latest version from https://github.com/google/material-design-icons/tree/master/variablefont
  - Also need to update `.ttf`, `.codepoint` and `.gitcommit`
  - Run `make generate-material-icons` again after update
  - Then run `make authui` for updating the UI
- [Twemoji Mozilla](authui/src/authflowv2/icons/Twemoji.Mozilla.woff2)
  - Download the latest versions from https://github.com/mozilla/twemoji-colr
  - Also need to update `.ttf` and `.gitcommit`
  - Run `make generate-twemoji-icons` again after update

## Generate translation

Scripts are located at `scripts/python/generate_translations.py`.

1. Add translation to your `base_language` `translation.json` file. `base_language` is defaulted as `en`

```diff
# resources/authgear/templates/en/translation.json
     "v2-error-phone-number-format": "Incorrect phone number format.",
+++  "v2-error-new-error": "Translate me"
```

2. Obtain your Anthropic api key. The translation is performed via `claude-3-sonnet-20240229` model

   If you are located in regions blocked by Anthropic, please make use of a VPN to access the holy Anthropic API.

3. Generate translations

```bash
make -C scripts/python generate-translations ANTHROPIC_API_KEY=<REPLACE_ME>
```

4. You should see

```log
python -m venv venv
...
2024-07-12 16:34:45,060 - INFO - ja | Translation result: {
  "v2-error-new-error": "私を翻訳する",
}
2024-07-12 16:34:45,060 - INFO - ja | Finished translation of chunk 1/1
2024-07-12 16:34:45,068 - INFO - ja | Updated ../../resources/authgear/templates/ja/translation.json with latest keys.
2024-07-12 16:34:45,068 - INFO - ja | Updating ../../resources/authgear/templates/ja/messages/translation.json with latest keys.
2024-07-12 16:34:45,069 - INFO - ja | Found 0 missing keys in ../../resources/authgear/templates/ja/messages/translation.json.
2024-07-12 16:34:45,071 - INFO - ja | Updated ../../resources/authgear/templates/ja/messages/translation.json with latest keys.
2024-07-12 16:34:45,071 - INFO - ja | Finished translation for ja (Japanese)
...
```

## Set up LDAP for local development

An openldap server and phpldapadmin has been set up in docker-compose.yaml for you already.

In case you need to do some LDAP related development, you need

### Create a LDAP user

- Go to http://localhost:18080. This is phpldapadmin
- You should see the login page. You need to sign in with the admin account.
  - The username is the environment variable `LDAP_ADMIN_DN`.
  - The password is the environment variable `LDAP_ADMIN_PASSWORD`.
- And then you need to create a group.
  - You create a group under the tree, indicated by `LDAP_ROOT`.
- And then you need to create a user.
  - You create a user under the tree, indicated by `LDAP_ROOT`. Assign the user to belong to the group you just created.

### Configure Authgear

In `authgear.yaml`, you add

```
authentication:
  identities:
  # Add this.
  - ldap
identity:
  # Add this.
  ldap:
    servers:
    - name: myldap
      url: ldap://localhost:1389
      base_dn: "dc=example,dc=org"
      user_id_attribute_name: "uid"
      search_filter_template: "(uid={{ $.Username }})"
```

In `authgear.secrets.yaml`, you add

```
- data:
    items:
    - name: myldap
      dn: "cn=admin,dc=example,dc=org"
      password: "adminpassword"
  key: ldap
```

### Start with the profile `ldap`

The ldap related services in `docker-compose.yaml` belong to the profile `ldap`.
To start them, you need to add `--profile ldap` to `docker compose up -d`, like

```
docker compose --profile ldap up -d
```

## Switching between sessionType=refresh_token and sessionType=cookie

The default configuration

- Accessing the portal at port 8000 or 8010
- AUTHGEAR_WEB_SDK_SESSION_TYPE in .env.example

assumes sessionType=refresh_token.

In case you need to switch to sessionType=cookie, you

- Use `AUTHGEAR_WEB_SDK_SESSION_TYPE=cookie` in your .env
- Access the portal at port 8001 or 8011

## Switch to Database config source

1. In your `.env` set these values:

```
CONFIG_SOURCE_TYPE=database
CUSTOM_RESOURCE_DIRECTORY=./var
```

2. Create a row in `_portal_config_source`:

```sh
go run ./cmd/portal internal configsource create ./var
```

3. Create domains for `accounts`:

```sh
# This allows portal to access the admin api with accounts.localhost
go run ./cmd/portal internal domain create-default --default-domain-suffix ".localhost"
# This allows using accounts.portal.localhost:3100
go run ./cmd/portal internal domain create-custom accounts --apex-domain="accounts.portal.localhost" --domain="accounts.portal.localhost"
# This allows using localhost:3100
go run ./cmd/portal internal domain create-custom accounts --apex-domain="localhost" --domain="localhost"
```

4. Create a free plan

This is needed if you want to create new projects.

```sql
INSERT INTO "public"."_portal_plan"("id","name","feature_config","created_at","updated_at")
VALUES
(E'free',E'free',E'{}',NOW(),NOW());
```

Restart your server, then it should be running with database config source.

# Storybooks


|        |                                                                                                                                                                                                     |
| ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Portal | [<img src="https://github.com/authgear/authgear-server/actions/workflows/chromatic.yaml/badge.svg?branch=ui-review-pending" />](https://ui-review-pending--67d2b9f5a864651d3793fe7e.chromatic.com/) |

# Agentic coding

We are trying out agentic coding workflow in this repository.

There are 2 types of reusable prompts:

- Always
- Manual

## Always prompts

The Always reusable prompts live in ./.cursor/.rules/

These rules have the following frontmatter at the beginning of the file

```
---
alwaysApply: true
---
```

When you add a new rule, you need to update ./CLAUDE.md to include that rule.

## Manual prompts

Another type of reusable prompts is Manual.

They have to be explicitly referenced by you.
They live in ./.claude/commands/

In Claude Code, they are custom slash commands, and you use them as such.

In Cursor, you need to reference them with `@`.
