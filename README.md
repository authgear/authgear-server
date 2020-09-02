# AuthGear
 
Work in progress

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
   go run ./cmd/authgear init config
   go run ./cmd/authgear init secrets
   ```

   then follow the instructions. For database URL and schema, use the following,
   ```
   DATABASE_URL=postgres://postgres@127.0.0.1:5432/postgres?sslmode=disable
   DATABASE_SCHEMA=app
   ```

6. Apply database schema migrations:

   make sure the db container is running

   ```sh
   go run ./cmd/authgear migrate up
   ```
   
To create new migration:
```sh
# go run ./cmd/authgear migrate new <migration name>
go run ./cmd/authgear migrate new add user table
```

## Comment tags

- `FIXME`: Should be fixed as soon as possible
- `TODO`: Should be done when someone really needs it.
- `OPTIMIZE`: Should be done when it really becomes a performance issue.
- `SECURITY`: Known potential security issue.