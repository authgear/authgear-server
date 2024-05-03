- [Contributing guide](#contributing-guide)
  * [Install dependencies](#install-dependencies)
  * [Set up environment](#set-up-environment)
  * [Set up the database](#set-up-the-database)
  * [Run](#run)
  * [Create an account for yourselves and grant you access to the portal](#create-an-account-for-yourselves-and-grant-you-access-to-the-portal)
  * [Known issues](#known-issues)
    + [Known issues on portal](#known-issues-on-portal)
  * [Comment tags](#comment-tags)
  * [Common tasks](#common-tasks)
    + [How to create a new database migration?](#how-to-create-a-new-database-migration)
    + [Set up HTTPS to develop some specific features](#set-up-https-to-develop-some-specific-features)
    + [Create release tag for a deployment](#create-release-tag-for-a-deployment)
    + [Keep dependencies up-to-date](#keep-dependencies-up-to-date)

# Contributing guide

This guide teaches you to run Authgear locally for development purpose.
It also covers some common development tasks.

## Install dependencies

This project uses asdf, and there is a .tool-versions file at the project root.

1. Install asdf
2. Run the following to install all dependencies in .tool-versions

   ```sh
   asdf install
   ```
3. Install icu4c

   ```sh
   brew install icu4c
   ```

   icu4c installed by brew is not globally visible by default, so you have to ensure your shell has the following in effect

   ```sh
   export PKG_CONFIG_PATH="$(brew --prefix)/opt/icu4c/lib/pkgconfig"
   ```

   To avoid doing the above every time you open a new shell, you may want to add it to your shell initialization script such as `~/.profile`, `~/.bash_profile`, etc.

4. Install libvips

   ```sh
   brew install vips
   ```

   libvips on macOS requires `-Xpreprocessor` to build.
   Run the following to tell Cgo.

   ```sh
   export CGO_CFLAGS_ALLOW="-Xpreprocessor"
   ```

5. Install libmagic

   ```sh
   brew install libmagic
   ```

   Run the following to tell Cgo where to find libmagic.
   Preferably you add it to your shell startup script.

   ```sh
   export CGO_CFLAGS="-I$(brew --prefix)/include"
   export CGO_LDFLAGS="-L$(brew --prefix)/lib"
   ```

6. Run `make vendor`.

## Set up environment

1. Set up environment variables

   ```sh
   cp .env.example .env
   ```

2. Generate config files

   ```sh
   go run ./cmd/authgear init -o ./var
   ```

   `authgear.yaml` must contain the following contents for the portal to work.

   ```yaml
   id: accounts # Make sure the ID matches AUTHGEAR_APP_ID environment variable.
   http:
      # Make sure this matches the host used to access main Authgear server.
      public_origin: "http://accounts.portal.localhost:3100"
   oauth:
      clients:
         # Create a client for the portal.
         # Since we assume the cookie is shared, there is no grant nor response.
         - name: Portal
           client_id: portal
           # Note that the trailing slash is very important here
           # URIs are compared byte by byte.
           redirect_uris:
              # This redirect URI is used by the portal development server.
              - "http://portal.localhost:8000/oauth-redirect"
              # This redirect URI is used by the portal production build.
              - "http://portal.localhost:8010/oauth-redirect"
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
           grant_types: []
           response_types:
              - none
   ```

3. Set up `.localhost`

   For cookie to work properly, you need to use

   - `portal.localhost:8000` to access the portal.
   - `accounts.portal.localhost:3100` to access the main server.

   You can either do this by editing `/etc/hosts` or using `dnsmasq`.

## Set up the database

1. Start the database

   ```sh
   docker compose up -d db
   ```

2. Apply migrations

   ```sh
   go run ./cmd/authgear database migrate up
   go run ./cmd/authgear audit database migrate up
   go run ./cmd/authgear images database migrate up
   go run ./cmd/portal database migrate up
   ```

## Run

1. In case you have made changes to authui, you run `make authui` to re-build the assets.
2. In case you have not started the dependencies, run `docker compose up -d`.
3. Run `make start` to run the main server.
4. Run `make start-portal` to run the portal server.
5. `cd portal; npm start` to run the portal frontend.

## Create an account for yourselves and grant you access to the portal

As the first project `accounts` is created by the script instead of by user, we need to add `owner` role to this project so as to gain access in the portal.

1. Register an account in [http://accounts.portal.localhost:3100](http://accounts.portal.localhost:3100)

   You will get the email otp code in the terminal which is running authgear server in the following form:

   ```
   skip sending email in development mode        app=accounts body="Email Verification\n\nThis email is sent to verify <your email> on Authgear. Use this code in the verification page.\n\n<your code>\n\nIf you didn't sign in or sign up please ignore this email.\n" logger=mail-sender recipient=<your email> reply_to= sender=no-reply@authgear.com subject="[Authgear] Email Verification Instruction"
   ```

   You can search this message with the keyword `Email Verification Instruction`.

2. Configure user permission for the project

   1. Connect to the database

   2. Go to the `_auth_user` table

   3. Copy the `id` value in the first row which is the account you registered

   4. Go to the `_portal_app_collaborator` table

   5. Create a new row of data

      1. For the `id` column, fill in with any string

      2. For the `app_id` column, fill in with `accounts`

      3. For the `user_id` column, fill in with the value you copied

      4. For the `created_at` column, fill in with `NOW()`

      5. For the `role` column, fill in with `owner`

      6. Save the data

   6. Now you can navigate to your project in the portal

## Known issues

cert-manager@v1.7.3 has transitive dependency problem.

siwe has to be 1.1.6. siwe@2.x has runtime error on page load. siwe@1.1.6 requires ethers@5.5.1.

### Known issues on portal

As `useBlocker` is removed since react-router-domv6.0.0-beta.7 and have no promise which version will
come back, we introduce the custom `useBlocker` hook by referencing the last commit which this hook
still exist.
See [https://github.com/remix-run/react-router/commit/256cad70d3fd4500b1abcfea66f3ee622fb90874](https://github.com/remix-run/react-router/commit/256cad70d3fd4500b1abcfea66f3ee622fb90874)
react-router-dom@6.4.0 removed the `block` function from NavigationContext.
We have to remain on react-router-dom@6.3.0 until we find an alternative.
As of react-router-dom@6.18.0, unstable_useBlocker and unstable_usePrompt are still marked as unstable.

@tabler/icons@1.92.0 is the last version that can be built with our current setup.
Newer version will cause our `npm run build` command to fail.

NPM has an outstanding issue related to optional native dependencies.
https://github.com/npm/cli/issues/4828
The issue will happen if the following conditions hold:
- The package.json, package-lock.json and node\_modules are in correct state. node\_modules only contain macOS arm dependencies.
- We update the version of parcel and run npm install to update package-lock.json
- package-lock.json becomes invalid.
- npm ci becomes broken on non macOS arm machines
So whenever we want to update dependencies, we first delete node\_modules and package-lock.json.
Then npm install will generate a correct package-lock.json.

When Parcel cannot resolve nodejs globals such as `process` and `Buffer`,
it installs them for us.
But we do not want to do that.
The workaround is to add `alias` to package.json.
See [https://github.com/parcel-bundler/parcel/issues/7697](https://github.com/parcel-bundler/parcel/issues/7697).

When we allow Parcel to perform tree shaking on code-splitted third party bundle,
refreshing a page will encounter module not found error.
To work around this, we disallow tree shaking in codesplit.ts.

Docker Desktop on Mac has [an issue](https://github.com/docker/for-mac/issues/5812#issuecomment-874532024) that would lead to an unresponsive reverse proxy.
One of the comment says enabling "Use the new Virtualization framework" would help.
After >5000 requests to the portal, "upstream timed out" errors will begin to pop up.
If enabling "Use the new Virtualization framework" did not help, you can restart Docker Desktop on Mac as a workaround.

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
make cert
# Uncomment the TLS config in nginx.confg to enable TLS, restart nginx to apply the change.
```

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
  - You can find out the server version by going into the container and run `apt list --installed`.
  - The version of Debian bookworm is [2024a](https://packages.debian.org/source/bookworm/tzdata), which correspond to [tzdata@v1.0.40](https://github.com/rogierschouten/tzdata-generate/releases/tag/v1.0.40).
- [The cropperjs type definition in ./authui/src](./authui/src)
- [GeoLite2-Country.mmdb](./GeoLite2-Country.mmdb)
- [GraphiQL](./pkg/util/graphqlutil/graphiql.go)
- [Material Icons](authui/src/authflowv2/icons/material-symbols-outlined.woff2)
  - Download the latest version from https://github.com/google/material-design-icons/tree/master/variablefont
  - Also need to update `.ttf`, `.codepoint` and `.gitcommit`
  - Run `make generate-material-icons` again after update
- [Twemoji SVG](authui/src/authflowv2/icons/twemoji-color.woff2), [Twemoji Mozilla](authui/src/authflowv2/icons/Twemoji.Mozilla.woff2)
  - Download the latest versions from https://github.com/13rac1/twemoji-color-font and https://github.com/mozilla/twemoji-colr
  - Also need to update `.ttf` and `.gitcommit`
  - Run `make generate-twemoji-icons` again after update
