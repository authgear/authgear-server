# AuthGear
 
Work in progress

## Troubleshooting

If you see an error `cannot change secret value` when you attempt to save secrets,
please make sure the content of `authgear.secrets.yaml` matches `APP_SECRET_*` environment variables.

## HTTPS setup

If you are testing external OAuth provider, you must enable TLS.

1. Cookie is only included in third party redirect if it has SameSite=None attribute.
2. Cookie with SameSite=None attribute without Secure attribute is rejected.

To setup HTTPS easily, you can use [mkcert](https://github.com/FiloSottile/mkcert)

```sh
# Install mkcert.
brew install mkcert
# Install the root CA into Keychain Access.
mkcert -install
# Create TLS certificate and private key with the given host.
mkcert -cert-file tls-cert.pem -key-file tls-key.pem localhost 127.0.0.1 ::1
```

One caveat is HTTP redirect to HTTPS is not supported, you have to type in https in the browser address bar manually.

## Prerequisite

Note that there is a local .tool-versions in project root. For the following setup to work, we need to

1. Install asdf

2. Run the following to install all dependencies in .tool-versions
   ```sh
   asdf install
   ```

3. Install icu4c

On macOS, the simplest way is to install it with brew

```sh
brew install icu4c
```

Note that by default icu4c is not symlinked to /usr/local, so you have to ensure your shell has the following in effect

```sh
export PKG_CONFIG_PATH="/usr/local/opt/icu4c/lib/pkgconfig"
```

To avoid doing the above every time you open a new shell, you may want to add it to your shell initialization script such as `~/.profile`, `~/.bash_profile`, etc.

## Database setup

1. Setup dependencies:
   ```sh
   make vendor
   ```
2. Setup environment variables (in `.env`):
   ```sh
   cp .env.example .env
   ```

3. start db container
   ```sh
   docker-compose up db
   ```

4. Create a schema:

   Run the following SQL command with command line to such as `psql` or DB viewer such as `Postico`

   ```sql
   CREATE SCHEMA app;
   ```

5. Initialize app

   To generate the necessary config and secret yaml file, run

   ```sh
   go run ./cmd/authgear init config --output ./var/authgear.yaml
   go run ./cmd/authgear init secrets --output ./var/authgear.secrets.yaml
   ```

   then follow the instructions. For database URL and schema, use the following,
   ```
   DATABASE_URL=postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable
   DATABASE_SCHEMA=app
   ```

6. Apply database schema migrations:

   make sure the db container is running

   ```sh
   go run ./cmd/authgear migrate up -f ./var/authgear.secrets.yaml
   ```
   
To create new migration:
```sh
# go run ./cmd/authgear migrate new <migration name>
go run ./cmd/authgear migrate new add user table
```

## Run server

To run development server, we need to start `db` and `redis` container

```sh
docker-compose up -d db redis
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

## Setup portal

Please refer to [Portal setup guide](./portal/README.md)

## Comment tags

- `FIXME`: Should be fixed as soon as possible
- `TODO`: Should be done when someone really needs it.
- `OPTIMIZE`: Should be done when it really becomes a performance issue.
- `SECURITY`: Known potential security issue.
