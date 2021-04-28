Error handling
==============

Error Kinds
-----------
The following kinds of error would appear in the codebase:

### Panic
Panics are produced with Go's built-in `panic` function. It is used for:
- impossible conditions (e.g. exhaustive enum matching)
- unrecoverable server error (e.g. invalid server/teanant configuration)

`panic` should be called with a `string` or `error` argument. If the argument
is a string, it should be in format `<module>: <message>`. For example:
`sso: unknown provider type: example`.

Panics are recovered by a middleware. Recovered panics would be logged and
produce a 500 response if needed.

### Plain error
Plain errors are produced by error constructing functions (e.g. `errors.New`)
or returned from dependency library. Unless otherwise required, this kind of
error should be used by default.

### API error
API errors are produced by `skyerr.Kind.NewXXX` series of functions. It should
be used for errors responding to client.

An API error consists of name, reason, message, and optionally info.

API error names correspond to a HTTP status code:
- BadRequest(400): server do not understand the request (i.e. syntactic error)
- Invalid(400): server understood the request, but refuse to process it (i.e. semantic error)
- Unauthorized(401): client do not have valid credentials (i.e. authentication error)
- Forbidden(403): client's credentials are not allowed for the request (i.e. authorization error)
- NotFound(404)
- AlreadyExists(409)
- TooManyRequest(429)
- InternalError(500)
- ServiceUnavailable(503)

Reason is an identifier indicating the kind of occurred error. Each reason
should be correspond to one error name. Reasons should be defined in the
error producing package, as opposed to centrally defined.

Message is a developer-facing message provided for ease of debugging. It should
not contain extra useful information not found in reason/info, in order to
prevent matching on the message by developers.

For errors requiring more detailed reporting (e.g. validation errors), an info
can be provided. Unless otherwise required, the info should contains a `cause`/
`causes` key, containing a cause object or array of cause object. Each cause
object should contains a string `kind` field.

A sample API error JSON:
```json
{
    "name": "Invalid",
    "reason": "PasswordPolicyViolated",
    "message": "password policy violated",
    "code": 400,
    "info": {
        "causes": [
            { "kind": "PasswordTooShort", "min_length": 8, "pw_length": 6 },
            { "kind": "PasswordUppercaseRequired" }
        ]
    }
}
```

### Error sentinel
Error sentinels are constant error value defined globally. The error value can
either be a plain error or API error. It is used to convey a well-known error
condition to the caller.


Error Wrapping
--------------
It is common to return a wrapped dependency error to caller.

### Simple wrapping
Use simple wrapped errors when adding context to errors from dependency:
```go
err := errors.Newf("failed to update user: %w", err)
```

### Secondary error
Use secondary error when error occured while handling error:
```go
err := errors.WithSecondaryError(err, errors.Newf("failed to rollback: %w", rerr))
```

### Detailed error
Use detailed error when additional information about the error can be provided:
```go
err := errors.WithDetails(err, errors.Details{"user_id": userID})
```

Detail values can be tagged to indicate the purpose of the value:
- `errors.SafeDetail`: details should be available to sysadmin only
- `skyerr.TenantDetail`: details should be available to tenant only
- `skyerr.APIErrorDetail`: details should be available to API client only
                           (as API error info entry)

For example, `sql` would be logged to console (i.e. visible to sysadmin):
```go
err := errors.WithDetails(err, errors.Details{"sql": errors.SafeDetail.Value(sql)})
```

Attached details can be collected using `errors.CollectDetails`.
### Error inspection
Wrapped error can be inspected using standard `errors.Is/As/Unwrap` API. For
API errors, `skyerr.IsKind` can be used to check if the error is specific kind.
Direct equality check against error sentinels is forbidden.

For logging purpose, a error summary (i.e. aggregated error messages) can be
produced using `errors.Summary`.

Architecture layers
-------------------
The system components can be roughly classified into three layers:

### Data layer
This layer handles IO to external services, such as database, email, and cloud
provider API. Most common components in this layer are stores.

Components in this layer should only return plain errors, or error sentinels;
they do not have sufficient context to produce a concrete API error.

### Logic layer
This layer handles the main application logic. Most common components in this
layer are providers.

Components in this layer should return plain errors, error sentinels, or API
errors. Errors from other logic layer components should be returned
directly or wrapped.

For some simple modules, a logic layer encapsulating component may be missing
(e.g. user profile). In this case, the component is treated as in logic layer
for error handling purpose.

### Transport layer
This layer handles IO to API requests. Most components in this layer are HTTP
handlers.

Components in this layer should return API errors. It is acceptable to
pass through errors from data/logic layer if the error is unexpected.

Errors returned by this layer would be converted to API errors.
For non API errors, an opaque internal error(i.e. 500) would be returned to
ensure no leakage of sensitive information.

Conventions
-----------

### Error definitions
Public errors in a package should be defined in single source file `error.go`.

For API error kinds (i.e. name & reason), it may be defined in the file for
easy usage:
```go
var UserNotFound = skyerr.NotFound.WithReason("UserNotFound")
```

For sentinel errors, the name should starts with `err`/`Err`:
```go
var ErrSessionNotFound = errors.New("session is not found")
```

### Database query result
If a query function returns a single entity (e.g. `GetUser(string) (*User, error)`),
it should return a sentinel error if the entity is not found.

If a query function returns multiple entities (e.g. `ListUsers() ([]*User, error)`),
it should return an empty slice without errors if the entity is not found.

### API error reason
API error reason should be defined such that:
- there exists consumers that would interested in the error at runtime
- consumer can react to the error reasonably
- consumer would not need to handle a large amount of reasons for single API

Examples:
- Having a `InvalidJSONBody` reason is not desirable: a correct app would not
  ever encounter this at runtime, and the app cannot recovered from this error.
  Instead, use a generic error reason through `skyerr.NewBadRequest`.
- Having a `VerifyCodeExpired` reason is not desirable: most consumer would
  only need to know whether the verification is successful. Instead, use a
  `UserVerificationFailed` reason, with a cause object with kind `ExpiredCode`.


References
----------
https://github.com/cockroachdb/errors
