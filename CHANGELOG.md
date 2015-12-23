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

