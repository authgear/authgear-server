# AuthGear

Work in progress

- [Prerequisite](#prerequisite)
- [Environment setup](#environment-setup)
- [Database setup](#database-setup)
- [HTTPS setup](#https-setup)
- [Running everything](#running-everything)
- [Multi-tenant mode](#multi-tenant-mode)
- [Portal setup](#portal-setup)
  - [Setup environment variable](#setup-environment-variable)
  - [Setup portal development server](#setup-portal-development-server)
- [Initial project setup](#initial-project-setup)
- [Known issues](#known-issues)
  - [Portal](#portal)
- [Comment tags](#comment-tags)
- [Credits](#credits)
- [Create release tag before deployment](#create-release-tag-before-deployment)
- [Keep dependencies up-to-date](#keep-dependencies-up-to-date)

## Prerequisite

Note that there is a local .tool-versions in project root. For the following setup to work, we need to

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

5. Run `make vendor`

## Environment setup

1. Setup environment variables:

   ```sh
   cp .env.example .env
   ```

2. Initialize app

   To generate the necessary config and secret yaml file, run

   ```sh
   go run ./cmd/authgear init -o ./var
   ```

3. Setup `.localhost` domain

   For cookie to work properly, you need to use

   - `portal.localhost:8000` to access the portal.
   - `accounts.portal.localhost:3100` to access the main server.

   You can either do this by editing `/etc/hosts` or install `dnsmasq`.

4. (Optional) To use db as config source.

   - Update `.env` to change `CONFIG_SOURCE_TYPE=database`
   - Setup config source in db
     ```
     go run ./cmd/portal internal setup-portal ./var/ \
        --default-authgear-domain=accounts.localhost \
        --custom-authgear-domain=accounts.portal.localhost \
     ```

## Database setup

1. Start the db container

   ```sh
   docker-compose up -d db
   ```

2. Apply database schema migrations:

   make sure the db container is running

   ```sh
   go run ./cmd/authgear database migrate up
   go run ./cmd/authgear audit database migrate up
   go run ./cmd/authgear images database migrate up
   go run ./cmd/portal database migrate up
   ```

To create new migration:

```sh
# go run ./cmd/authgear database migrate new <migration name>
go run ./cmd/authgear database migrate new add user table
```

## HTTPS setup

TLS is required to perform OAuth and WebAuthn.

To setup HTTPS easily, you can use [mkcert](https://github.com/FiloSottile/mkcert)

```sh
# Install mkcert.
brew install mkcert
# Install the root CA into Keychain Access.
mkcert -install
# Generate the TLS certificate
make mkcert
# Uncomment the TLS config in nginx.conf to enable TLS
# Restart proxy
```

One caveat is HTTP redirect to HTTPS is not supported, you have to type in https in the browser address bar manually.

## Running everything

```sh
docker-compose up -d
```

Then run the command

```sh
# in project root
go run ./cmd/authgear start
```

To run graphql server

```sh
# in project root
go run ./cmd/portal start
```

## Multi-tenant mode

Some features (e.g. custom domains) requires multi-tenant mode to work properly.
To setup multi-tenant mode:

1. Setup local mock Kubernetes servers:
   ```
   cd hack/kube-apiserver
   docker-compose up -d
   ```
2. Install cert manager CRDs:

   ```
   kubectl --kubeconfig=hack/kube-apiserver/.kubeconfig apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v1.3.1/cert-manager.crds.yaml
   ```

3. Bootstrap Kubernetes resources:

   ```
   kubectl --kubeconfig=hack/kube-apiserver/.kubeconfig apply -f hack/k8s-manifest.yaml
   ```

4. Enable multi-tenant mode in Authgear & portal server:
   refer to `.env.example` for example environment variables to set

## Portal setup

### Setup environment variable

We need to set up environment variables for Authgear servers and portal server.

Make a copy of `.env.example` as `.env`, and update it if necessary.

### Setup portal development server

1. Install dependencies

   ```
   npm ci
   ```

2. Run development server

   ```
   npm start
   ```

   This command should start a web development server on port 1234.

3. Configure authgear.yaml

   We need the following `authgear.yaml` to setup authgear for the portal.

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

## Initial project setup

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

### Portal

As `useBlocker` is removed since react-router-domv6.0.0-beta.7 and have no promise which version will
come back, we introduce the custom `useBlocker` hook by referencing the last commit which this hook
still exist.
See [https://github.com/remix-run/react-router/commit/256cad70d3fd4500b1abcfea66f3ee622fb90874](https://github.com/remix-run/react-router/commit/256cad70d3fd4500b1abcfea66f3ee622fb90874)
react-router-dom@6.4.0 removed the `block` function from NavigationContext.
We have to remain on react-router-dom@6.3.0 until we find an alternative.

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

## Credits

- Free email provider domains list provided by: https://gist.github.com/tbrianjones/5992856/
- This product includes GeoLite2 data created by MaxMind, available from [https://www.maxmind.com](https://www.maxmind.com)

## Create release tag before deployment

```sh
# Create release tag when deploying to staging or production
# For staging, prefix the tag with `staging-`. e.g. staging-2021-05-06.0
# For production, no prefix is needed. e.g 2021-05-06.0
# If there are more than 1 release in the same day, increment the last number by 1
git tag -a YYYY-MM-DD.0

# Show the logs summary
make logs-summary A=<previous tag> B=<current tag>
```

## Keep dependencies up-to-date

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
  - The version of Debian bullseye is `2021a`, which correspond to `tzdata@v1.0.25`.
- [The cropperjs type definition in ./authui/src](./authui/src)
- [GeoLite2-Country.mmdb](./GeoLite2-Country.mmdb)
- [GraphiQL](./pkg/util/graphqlutil/graphiql.go)
