## 0.7.0 (2016-03-02)

### Features

- Check record permission on record:* handler
- Add CloudFormation template and scripts #561
- Add ACL matching by JSON pattern #299

## 0.6.0 (2016-02-24)

### Features

- Pass skygear configuration to plugin #492
- Return 403 Forbidden for PermissionDenied error
- Implement saving role based acl #298
- Add checking of user permission before modify roles #539
- Only init db for the first time db is opened #573
- Update build script to build with zmq in linux
- Bring back subscription:{save,fetch}
- Remove the alembic migrate related deployment files #268
- Skygear will migrate db itself #268
- Reject request with duplicated roles specified user #564
- Support for master key #294

### Bug Fixes

- Fix bug on serizlise Sequence to plugin #559
- Fix bug on base64 encode at exec transport

### Other Notes

- Add go generate sources #571
- Add script for cross compile binaries

## 0.5.0 (2016-02-17)

### Features

- Make zmq optional and it is not compiled by default #543
- Implement `schema:*` handler for getting and modifying database schema #491
  - `schema:fetch`
  - `schema:create`
  - `schema:rename`
  - `schema:delete`
- Add middleware to support CORS #273
- Implement http transport and support request context to console transport #537, #538
- Better exec transport by providing config and print log in skygear #538
- Add `role:admin` and `role:default` for system config #295
- Support for user record #409
  Skygear will create user record that have same id as `_user` when user signup.
  - Behaviour of `auth:login` and `auth:signup` is modified to create user record
    when a new user is created.
  - Behaviour of user:query and relation:query remains unchanged.
  - `record:query` is extended to support `UserDiscoverFunc` which returns
    user by email address.
  - `record:query` returns user record when eager loading user through
    reserved fields (e.g. `_owner_id`).
  - It is not allowed to delete user record.
- Add `DevOnlyProcessor` to restrict dev-only endpoint
- Pass plugin exception info to client oursky/py-skygear#109
- Support registering multiple hooks of same kind oursky/py-skygear#108

### Bug Fixes

- Fix dev_only preprocessor wrongly required by home handler #549
- Fix zmq socket leak #425, #527

### Other Notes

- Update docker-compose.yml to version 2
- Unify handler to use mapstructure to convert the payload #545
- Update goczmq

## 0.4.1 (2016-01-28)

### Features

- Implement updates of user roles via user:update #296, #295

### Bug Fixes
- Fix serializing a wrong location field to plugin #519
- Recover from zmq crash and log to errors #527
- Fix before save hook without ownerID #528

### Other Notes
- Declare preprocessors by dependency injection #499
- Make the Processor an interface with Preprocess func #501
- Unify handler and plugin serialization #519

## 0.4.0 (2016-01-13)
### Features

- Request context is now passed from skygear to plugin. Only lambda and hook
  are supported #470
- Lambda function can specify whether authenticated user or access key is
  required #267

### Other Notes

- Refractor handler as struct and use facebookgo/inject to manage dependency
  #482
- Specify access control type through configuration #297

## 0.3.0 (2016-01-06)

### Features

- Show executed SQL count in log #428
- Include signed url on asset uploaded response #427
- User relation query now supports pagination using offset and limit
  parameter #456

### Bug Fixes

- Panic is now catched by router and appropriate response returned #478
- Status code for some error condition
- Removed fs database driver #433
- Incorrect error code when changing password #408
- Properly log Plugin transport state changes #279
- Return status OK on logout success
- Make public database as default database
- Panic when trying to logout a user #477
- Improve reliability for zmq plugin during init #453, #452
- Fix skygear fail to start because cert/key path cannot be opened, even if
  APNS is disabled #461
- Fix not terminating coroutine after websocket connection has closed
- Fix unable to send push notifications to all devices when multiple are
  configued #462
- Retry plugin init request rather than waiting indefinitely #452
- Deduplicate the device.Token to send to user devices
- Send to all deivces of a user instead of last device at push to user handler

### Other Notes

- Temporarily require only naive API key for asset upload #470

## 0.2.0 (2015-12-23)

### Features

- Require authenticated user for subscription and device register #431
- Support public asset store, return an un-signed URL for public store #385
- Better error detection when query is malformed, especially when comparing
  map with keypath #339
- Introduce consistent error code #427
- Eager load records in a batch using SQL `IN` operator #395

### Bug Fixes

- Retry opening connection to database when starting #440
- Fix bug on transient field returning a wrong object #436
- Fix unable to upload asset with `+` in file name #426

