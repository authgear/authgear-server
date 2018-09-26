![Skygear Logo](.github/skygear-logo.png)

Next is the V2 of Skygear that aim to follow

- Support multi tenant at core, make cloud deploy scalable at first day.
- Simplify deployment.
- Give back application lifecycle to cloud code developer and skygear
  developer.
- Drop zmq and model against HTTP semantics.

## Project structure

```
.
├── pkg
│   ├── server    <-- original skygear-server code
│   ├── auth
│   ├── record
│   ├── gateway
│   └── core
│
└── cmd
    ├── auth
    │   └── main.go
    │
    ├── record
    │   └── main.go
    │
    └── gateway
        └── main.go
```

## Gateway

### Migration

To provide db url by `MIGRATION_DB=`, default is `postgres://postgres:@localhost/postgres?sslmode=disable` for developement

**Add a version**

```
cd ./scripts/gateway
$ make add-version REVISION=<revision_description>
```

**Run migration**

```
cd ./scripts/gateway
$ make migrate-up
```

**Check current db version**

```
cd ./scripts/gateway
$ make migrate-version
```

### To add a new gear

1. Add migration to update plan table with gear enabled column
2. Update `Plan` struct and `CanAccessGear` func in `pkg/gateway/model/plan.go`
3. Update `RouterConfig` and `GetRouterMap` func in `pkg/gateway/config/config.go`

## Gear DB migration

The following part is about gear db migration.

If you come from skygear-server 0.x to 1.x, the biggest difference is that skygear next would not offer JIT db migration in most gears.

DB migration must be run before server boot up. And since we do not have a full featured db management tool for skygear yet, here is a general guide for new comers of skygear next user.

1. Create a schema for your app.
1. Run core migration.
1. Run gear(s) migration.

*See previous section if you also need to run db migration for gateway.*

For example, the app name is `helloworld` and you want to run `auth` gear and `chat` gear.

```
# Create a schema for your app, with name app_{name}
# Run the following SQL in any postgresql client, like Postico
CREATE SCHEMA app_helloworld;

# If you have psql cli
$ psql ${DATABASE_URL} -c "CREATE SCHEMA app_helloworld;"

# Run core migration
$ go run cmd/migrate/main.go -path cmd/migrate/gear/core -schema app_helloworld -gear core up

# Run auth gear migration
$ go run cmd/migrate/main.go -path cmd/migrate/gear/auth -schema app_helloworld -gear auth up

# Run chat gear migration
$ go run cmd/migrate/main.go -path cmd/migrate/gear/chat -schema app_helloworld -gear chat up
```

See below sections for more commands about db migration.

### Commands

**Add a version**

```
$ export GEAR=<gear_name>
$ export REVISION=<revision_description>
$ make -f scripts/gateway/Makefile add-version MIGRATION_DIR=cmd/migrate/gear/${GEAR} REVISION=${REVISION}
```

**Run core migration**

```
$ export APPNAME="app name"
$ go run cmd/migrate/main.go -path cmd/migrate/core -schema app_${APPNAME} -core up 1
```

**Run gear migration**

```
$ export GEAR="gear name"
$ export APPNAME="app name"
$ go run cmd/migrate/main.go -path cmd/migrate/gear/${GEAR} -schema app_${APPNAME} -gear ${GEAR} up 1
```

**Check current db version**

```
$ export GEAR="gear name"
$ export APPNAME="app name"
$ go run cmd/migrate/main.go -path cmd/migrate/gear/${GEAR} -schema app_${APPNAME} -gear ${GEAR} version
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
