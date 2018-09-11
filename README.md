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

## License & Copyright

```
Copyright (c) 2015-present, Oursky Ltd.
All rights reserved.

This source code is licensed under the Apache License version 2.0
found in the LICENSE file in the root directory of this source tree.
An additional grant of patent rights can be found in the PATENTS
file in the same directory.

```
