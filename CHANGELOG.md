## 0.23.0 (2017-04-20)

### Features

- Implement GeoJSON query using ST_Contains (#262)
- Make default ACL server-based (#309)

### Bug Fixes

- Ensure respond channel of zmq is always consumed so channeller is not blocked (#347)
- Fix inconsistent behavior saving to integer column (#319)

### Other Notes

- Revamp query sqlizer into its own package

## 0.22.2 (2017-03-31)

### Features

- Create user when master key is provided (#261)
- Fetch full assets data in query
- Return asset content type in response of record save

### Bug Fixes

- Fix plugin timer request not properly initialized (#333)
- Fix schema:fetch when record type has no fields


### Other Notes

- Make RequireUser preprocessor gives a better error message
- Add ASSET_STORE_S3_URL_PREFIX in .env.sample
- Update godev image not using development version of glide
- Commands for setting up skygear in Linux env

## 0.22.1 (2017-02-15)

### Bug Fixes

- Fix unable to establish pubsub connection because of CloseNotifier (#291)

## 0.22.0 (2017-02-10)

### Features

- Implement API response timeout (#271)

### Bug Fixes

- Fix request context not initialized (#284)
- Check for constraint violation when deleting

### Other Notes

- Add Nix derivation for building
- Require go1.7.4

## 0.21.1 (2017-01-19)

### Features

- Add `_status:healthz` endpoint (#264)

### Bug Fixes

- Fix ST_DWithin query (#266)

## 0.21.0 (2017-01-11)

### Features

- Support token based APNS (#239)

## 0.20.0 (2016-12-20)

### Features

- Support s3 asset store url prefix (#254)
- Support unregister device (#249)
- Add configuration for timeout of plugin
- Provide skygear config for plugin init events

### Bug Fixes

- Fix postgis-2.3 not found

### Other Notes

- More log when plugin response is invalid (#232)
- Put access_key_type to plugin request context (#238)
- Limit the number of bytes in request log

    The existing implementation prints the response body for certain type
    this pose a problem to log collection when the response body is too big.

- Better handling for unrecognized data type (#231)

    This is achieved by introducing a new skydb type called Unknown, which
    will be passed to client when the server sees a column with an
    unrecognized data type.

    Previous behavior will result in skygear server throwing an error when
    it sees a column with unrecognized data type.

## 0.19.1 (2016-11-12)

### Bug Fixes

- Fix plugin lambda / handler always require API key

## 0.19.0 (2016-11-10)

### Incompatible changes

- The protocol for plugin transport is updated. Skygear Server and cloud code
  in previous versions cannot be used with this version.

### Features

- Support plugin only request during plugin initialization
- Add support for bigint db type
- Make use of `ST_DWithin` to compare distance (#213)
- Support plugin event (#199)

### Bug Fixes

- Fix various issues with creating user on sign up (#218)
- Fix schema extending more than necessary
- Fix unable to configure sentry log hook

### Other Notes

- Send JSON-encoded plugin call in http request body (skygeario/py-skygear#82)

## 0.18.0 (2016-10-28)

### Features

- Refresh token at me endpoint (#118)

### Bug Fixes

- Fix unable to detect schema conflict (#140)
- Check for username/email duplicate when updating (#124)
- Resolve socket exhaust problem on high concurrency (#160)
- Fix skygear not sending init to plugin when restarted quickly (#150)
- Fix auth:password response invalid access token (#142)

### Other Notes

- Properly stop pubsub hub at test case
- Allow run test with docker-compose (#157)
- Make changes for smaller number of image layers
- Add ca-certificate to deps building pipeline (#154)
- Mark the skygear-server restart to always in Docker Compose (#145)

## 0.17.1 (2016-09-23)

### Bug Fixes

- Fix not able to update user after signup with auth provider

## 0.17.0 (2016-09-15)

### Bug fixes

- `auth:login` return last seen that query from DB,
  not current timestamp (#110)
- Fix bugs on `me` clear out the last login at for user

## 0.16.0 (2016-09-02)

### Features

- Implement slave mode (#103, #104, #105)
- Provide user last logged in and last seen at _user
- Support Cloud Asset (#107)
- Implement `me` endpoint to get current user (#111)

### Bug Fixes

- Update amz.v3 for signing s3 asset with UTC

### Other Notes

- Revamp release binaries building
- Upgrade postgres to 9.5

## 0.15.0 (2016-08-17)

### Features

- Schema migration is disable in non-dev-mode (#93)
- Add migration for initial admin user (#75)
- Assign default user role to new signup user (#44)
- Support user discovery with username (#19)

### Bug Fixes

- Fix issue when running `go install` (#64)
- Remove existing device with the same token when registering device (#71)
- Fix JWT token not considered valid created by signup (#94)

### Other Notes

- Update setup test env. script on README

## 0.14.0 (2016-07-26)

### Features

- Implement predicate with keypath to referenced record (#85)
- Set default log level of plugin logger to INFO (#49)
- Add JWT token store (#74)
- Include user role in user query (#70)
- Update to use HTTP/2 APNS protocol (#47)

### Bug fixes

- Create extensions when migrating database schema (#53)
- Handle invalid data format for pubsub actions (SkygearIO/skygear-SDK-JS#27)
- Check field exists before performing query (#6)
- Fix scan NULL token in QueryDevicesByUser (#33)
- Preserve ACL when saving record with ACL=nil (#38)
- Fix not-predicate not sqlized (#78)
- Fix Public record is accessible without userinfo

### Other Notes

- Config auth token expiry time and default not to expire (#65)

## 0.13.0 (2016-07-05)

### Features

- Allow user to add role with master key
- Implement union database, which contains all records across public
  and private databases, only accessible by client with master key
- Bypass access control with master key (#51)

### Bug fixes

- Fix ACCESS_CONTROL config default
- Fix ACL incorrectly bypassed in certain condition (#58)

### Other Notes

- Make the CORSHost default to `*`
- Switch to go 1.6

## 0.12.1 (2016-06-02)

### Bug fixes

- Read correct getSentry log level from ENV VAR (#43)
- Read GCM config from env var

## 0.12.0 (2016-05-30)

### Incompatible changes

- Read all config from ENVVAR and support .env files (#35)

### Bug fixes

- Make `_user` email/username to be case insensitive at pq (#41)
- Fix the public read record ACL bug on non readable (#39)
- Consider deleting non-existing device as success

### Other Notes

- Update the doc link to http://docs.skygear.io/ (#37)
- Update travis build status badge

## 0.11.0 (2016-05-09)

### Features
- Allow master key to override ACL restriction (#22)

### Bug Fixes
- Check sequence exist before update integer columns (#6)
- Fix missing headers returned from plugins (#15)
- Fix travis build error on Go 1.6

### Other Notes
- Use mime package for mime processing and allow config of mime type concern (#25)
- Update slack notification token
- Update quickstart example (oursky/skygear-doc#162)

## 0.10.0 (2016-04-13)

### Features
- Add version number to getsentry event (oursky/skygear-server#624)
- Support public read write ACL (oursky/skygear-server#647)
- Allow use of arbitrary HTTP method name (oursky/py-skygear#135)
- Add server version on log and request header (oursky/skygear-server#623)

### Bug Fixes
- Add Checking whether auth provider exists (skygeario/skygear-server#3)
- Fix unable to query keypath for null (oursky/skygear-server#635)
- Fix last subscriber stealing all published message (oursky/skygear-server#642)

## 0.9.0 (2016-03-16)

### Features

- Support record access control on creation by role #594
- Accept env SKY_CONFIG as config filepath #605
- Implement Handler provided by plugin #587
- Add key prefix to redis token store #616

### Bug Fixes

- Retry http plugin init until success #598

## 0.8.0 (2016-03-09)

### Features

- Add HTTP path routingto router #90
- Support quickstart example with plugin deploy

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
