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

