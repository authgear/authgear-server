![Skygear Logo](.github/skygear-logo.png)

Next is the V2 of Skygear that aim to follow

- Support multi tenant at core, make cloud deploy scalable at first day.
- Simplify deployment.
- Give back application lifecycle to cloud code developer and skygear
  developer.
- Drop zmq and model against HTTP semantics.
- Drop v1 Record class

## Project structure

```
.
├── pkg
│   ├── server    <-- original skygear-server code
│   ├── auth
│   ├── gateway
│   └── core
└── cmd
    ├── auth
    │   └── main.go
    └── gateway
        └── main.go
```

## Dependencies

If you plan to build and run locally, you need to install the following dependencies.

- pkgconfig
- vips >= 8.7

If you are on macOS and user of homebrew, you can install them by

```sh
brew install pkgconfig vips
```

## Gateway

### Migration

- See [DB migration](#db-migration), use `gateway` for `module`.

### To add a new gear

1. Add db migration to config db
    - Add enabled column to plan table
    - Add version column to app table
1. Update `pkg/gateway/model/app.go` with new gear in `Gear`, `App` and `GetGearVersion`.
1. Update `GetAppByDomain` in `pkg/gateway/db/app.go` with the new gear version column.
1. Update `Plan` struct and `CanAccessGear` func in `pkg/gateway/model/plan.go`
1. Update `GearURLConfig` and `GetGearURL` func in `pkg/gateway/config/config.go`

## DB migration

The following part is about gateway and gears db migration.

If you come from skygear-server 0.x to 1.x, the biggest difference is that gears in skygear next would not support auto db migration in server boot time.

DB migration must be run before server boot up. And since we do not have a full featured db management tool for skygear yet, here is a general guide for new comers of skygear next user.

1. Create a schema for common gateway.
1. Create a schema for your app.
1. Run core and gear(s) migration.

For example, the app name is `helloworld` and you want to run `auth` gear .

```
# Base app_config schema for core gateway
CREATE SCHEMA app_config;
# Create shared schema for apps
# Run the following SQL in any postgresql client, like Postico
CREATE SCHEMA app;

# If you have psql cli
$ psql ${DATABASE_URL} -c "CREATE SCHEMA app;"

# Run core and auth migration
$ make -C migrate migrate MIGRATE_CMD=up DATABASE_URL=${DATABASE_URL} SCHEMA=app
```

See below sections for more commands about db migration.

### Commands

**Add a version**

```sh
# MODULE can be gateway, core, auth...
$ export MODULE=<module_name>
$ export REVISION=<revision_description>
$ make -C migrate add-version MODULE=${MODULE} REVISION=${REVISION}
```
**Check current db version**

```
$ make -C migrate
```

**Dry run the migration**

Transaction will be rollback

```
$ make -C migrate MIGRATE_CMD=up DRY_RUN=1
```

**Run the migration with github source**

```
$ make -C migrate migrate \
    CORE_SOURCE=github://:@skygeario/skygear-server/migrations/core#6918eed \
    AUTH_SOURCE=github://:@skygeario/skygear-server/migrations/auth#6918eed \
    MIGRATE_CMD=up
```

**Running db migration to all apps in cluster (multi-tenant mode)**

Run core and auth migrations to apps which auth version in live

```
$ make -C migrate migrate \
    APP_FILTER_KEY=auth_version \
    APP_FILTER_VALUE=live \
    CONFIG_DATABASE=postgres://postgres:@localhost/postgres?sslmode=disable \
    HOSTNAME_OVERRIDE=localhost \
    MIGRATE_CMD=up
```

**Start migration server in http server mode**

- To start the migration server

```
$ make -C migrate http
```

- Calling the migration server 

    ```
    POST /migrate
    ```

    Request example

    ```json
    {
        "migration": "auth",
        "schema": "app_config",
        "database": "postgres://postgres:@localhost:5432/postgres?sslmode=disable",
        "command": "version"
    }
    ```

    Response example

    ```json
    {
        "result":"1563434450"
    }
    ```

## License & Copyright

```
Copyright (c) 2015-present, Oursky Ltd.
All rights reserved.

This source code is licensed under the Apache License version 2.0
found in the LICENSE file in the root directory of this source tree.
An additional grant of patent rights can be found in the PATENTS
file in the same directory.

```
