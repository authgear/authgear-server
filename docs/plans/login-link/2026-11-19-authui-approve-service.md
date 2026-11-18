# Auth UI approves login links through the login link API — implementation plan

Spec: [login-link.md](../../specs/login-link.md). Follows [2026-11-18-custom-ui-login-link.md](./2026-11-18-custom-ui-login-link.md).

## 1. Goal / scope

Everything the Auth UI approval page does must be possible through the login link API, so a Custom UI can build the same page. So the HTTP API and Auth UI call the same request-level methods: an `APIController` takes a request body, validates it with the endpoint's request schema, and passes the code to `loginlink.Service`. Doing so exposes a gap: Auth UI checks a link on page load, and the API had no way to do that.

In scope:

- New `POST /api/v1/login_link/inspect`: the same request, response and errors as `approve`, without approving (spec: [API](../../specs/login-link.md#api)).
- New `pkg/lib/feature/loginlink.Service` with `Inspect(ctx, code)` and `Approve(ctx, code)`.
- New `pkg/auth/api.APIController` with `InspectLoginLink(ctx, body)` and `ApproveLoginLink(ctx, body)`.
- The `inspect` and `approve` HTTP handlers read the JSON body and call `APIController`.
- The Auth UI approval page (`pkg/auth/handler/webapp/authflowv2/verify_login_link.go`) calls `APIController` with a body it builds, and no longer calls the OTP service.

Out of scope:

- Latte (`pkg/latte/intent_verify_login_link.go`), the legacy NWD UI.
- The page's HTML, query states (`""`, `invalid_code`, `matched`) and form field `x_oob_otp_code`.

## 2. Config model and schema

No config change.

## 3. Runtime flow

### 3.1 `pkg/lib/feature/loginlink`

`service.go`:

```go
package loginlink

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package loginlink

type OTPCodeService interface {
	LookupCode(ctx context.Context, purpose otp.Purpose, code string) (target string, err error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error)
}

type EventStore interface {
	Publish(ctx context.Context, websocketChannelName string, e authflow.Event) error
}

type Service struct {
	Config   *config.AppConfig
	OTPCodes OTPCodeService
	Events   EventStore
}

// Inspect returns nil when code can be approved.
func (s *Service) Inspect(ctx context.Context, code string) error

// Approve approves code and refreshes the Authentication Flow that sent it.
func (s *Service) Approve(ctx context.Context, code string) error

func (s *Service) verify(ctx context.Context, code string) (kind otp.Kind, target string, err error)
```

`deps.go`: `var DependencySet = wire.NewSet(wire.Struct(new(Service), "*"))`.

Call flow (today's logic, moved from the API handler):

- `verify`:
  1. `kind := otp.KindOOBOTPLink(s.Config, model.AuthenticatorOOBChannelEmail)`.
  2. `target, err := s.OTPCodes.LookupCode(ctx, kind.Purpose(), code)`. Unknown or expired: `InvalidOTPCode`.
  3. `s.OTPCodes.VerifyOTP(ctx, kind, target, code, &otp.VerifyOptions{SkipConsume: true})`. Consumed, or replaced by a resend: `InvalidOTPCode`. Limits: `RateLimited`. Nothing is consumed.
- `Inspect`: `_, _, err := s.verify(ctx, code)`; return `err`.
- `Approve`: `kind, target, err := s.verify(ctx, code)`, then:
  1. `state, err := s.OTPCodes.SetSubmittedCode(ctx, kind, target, code)`.
  2. If `state.AuthenticationFlowWebsocketChannelName != ""`, publish `authflow.NewEventRefresh()` to it.
  3. Return nil.

Wiring, `pkg/lib/deps/deps_common.go`, next to `authenticationflow.DependencySet` (which provides `*authenticationflow.WebsocketEventStore`; `*otp.Service` is provided in the same file):

```go
loginlink.DependencySet,
wire.Bind(new(loginlink.OTPCodeService), new(*otp.Service)),
wire.Bind(new(loginlink.EventStore), new(*authenticationflow.WebsocketEventStore)),
```

### 3.2 `APIController`

New `pkg/auth/api/api_controller.go`. It is the request-level entry of the API, called by the HTTP handlers and by Auth UI. `pkg/auth/api` is the API-side counterpart of `pkg/auth/webapp`, and imports only `pkg/api`, `pkg/lib` and `pkg/util`. So the HTTP handlers do not depend on Auth UI, and Auth UI depends on the API, not the other way round. Its package name `api` clashes with `pkg/api`, so importers alias it `authapi`.

```go
//go:generate go tool mockgen -source=api_controller.go -destination=api_controller_mock_test.go -package api

// LoginLinkV1RequestSchema is the request body of POST /api/v1/login_link/inspect
// and POST /api/v1/login_link/approve. Moved from login_link_v1_approve.go.
var LoginLinkV1RequestSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false,
		"properties": {
			"code": { "type": "string", "minLength": 1 }
		},
		"required": ["code"]
	}
`)

type APIControllerLoginLinkService interface {
	Inspect(ctx context.Context, code string) error
	Approve(ctx context.Context, code string) error
}

type APIController struct {
	LoginLinks APIControllerLoginLinkService
}

// InspectLoginLink is POST /api/v1/login_link/inspect. body is the request body.
func (c *APIController) InspectLoginLink(ctx context.Context, body map[string]any) error

// ApproveLoginLink is POST /api/v1/login_link/approve. body is the request body.
func (c *APIController) ApproveLoginLink(ctx context.Context, body map[string]any) error

func (c *APIController) parseLoginLinkV1Code(ctx context.Context, body map[string]any) (string, error)
```

- `parseLoginLinkV1Code`:
  1. `LoginLinkV1RequestSchema.Validator().ValidateValueWithMessage(ctx, body, "invalid request body")`. On failure, return the `ValidationFailed` error. A nil map (a JSON `null` body) fails as not an object.
  2. Return `body["code"].(string)`; the schema guarantees a non-empty string. No further JSON encoding or decoding.
- `InspectLoginLink`: `code, err := c.parseLoginLinkV1Code(ctx, body)`; then `c.LoginLinks.Inspect(ctx, code)`.
- `ApproveLoginLink`: the same, with `c.LoginLinks.Approve`.

Wiring:

- `pkg/auth/api/deps.go`: `wire.Struct(new(APIController), "*")`, in the existing `DependencySet`.
- `pkg/auth/deps.go`: `wire.Bind(new(authapi.APIControllerLoginLinkService), new(*loginlink.Service))`.

### 3.3 HTTP handlers

`login_link_v1_approve.go`:

- Remove `LoginLinkV1ApproveRequestSchema`, `LoginLinkV1ApproveRequest`, `LoginLinkV1ApproveOTPCodeService`, `LoginLinkV1ApproveEventStore` and `approve`.
- `LoginLinkV1ApproveHandler` becomes `{ Controller *authapi.APIController }`.
- `ServeHTTP`:
  1. `var body map[string]any`; `httputil.ParseJSONBody(r, w, func(reader io.Reader, value any) error { return json.NewDecoder(reader).Decode(value) }, &body)`. `ParseJSONBody` keeps today's content type check and size limit; the handler does not validate.
  2. `err = h.Controller.ApproveLoginLink(ctx, body)`.
  3. Errors are logged and written as today; on success, `api.Response{Result: struct{}{}}`.

New `login_link_v1_inspect.go`: `ConfigureLoginLinkV1InspectRoute` (`OPTIONS`, `POST`, `/api/v1/login_link/inspect`) and `LoginLinkV1InspectHandler { Controller *authapi.APIController }`, the same as approve but calling `InspectLoginLink`, with its own logger `handler-login-link-v1-inspect`.

Wiring:

- `pkg/auth/routes.go`: `router.Add(apihandler.ConfigureLoginLinkV1InspectRoute(authenticationFlowRoute), p.Handler(newAPILoginLinkV1InspectHandler))`, next to approve.
- `pkg/auth/wire_handler.go`: `newAPILoginLinkV1InspectHandler`.
- `pkg/auth/handler/api/deps.go`: `wire.Struct(new(LoginLinkV1InspectHandler), "*")`.
- `pkg/auth/deps.go`: remove the binds for `handlerapi.LoginLinkV1ApproveEventStore` and `handlerapi.LoginLinkV1ApproveOTPCodeService`.

### 3.4 Auth UI approval page

`verify_login_link.go`:

- `AuthflowV2VerifyLoginLinkOTPHandler` keeps `ControllerFactory` (for `ctrl.Get` and `ctrl.PostAction`), `BaseViewModel`, `AuthenticationViewModel` and `Renderer`, and gets `APIController *authapi.APIController`. Remove `LoginLinkOTPCodeService`, `GlobalSessionServiceFactory`, `AuthenticationFlowEvents`, `Config`, and the `AuthenticationFlowWebsocketEventStore` interface.
- `ctrl.Get`, initial state:
  1. `code := r.URL.Query().Get("code")`. If `code == ""`, `finishWithState(LoginLinkOTPPageQueryStateInvalidCode)`. Today an empty code fails the lookup and shows this state; the schema would otherwise make it a validation error page.
  2. `err := h.APIController.InspectLoginLink(ctx, map[string]any{"code": code})`.
  3. `InvalidOTPCode` → `finishWithState(LoginLinkOTPPageQueryStateInvalidCode)`; another error → return it.
- `ctrl.PostAction("")`:
  1. Validate the form and read `x_oob_otp_code`, as today. The form schema requires the field but allows `""`, so if `code == ""`, `finishWithState(LoginLinkOTPPageQueryStateInvalidCode)`, as today's lookup miss does.
  2. `err := h.APIController.ApproveLoginLink(ctx, map[string]any{"code": code})`.
  3. `InvalidOTPCode` → `finishWithState(LoginLinkOTPPageQueryStateInvalidCode)`; another error → return it.
  4. `finishWithState(LoginLinkOTPPageQueryStateMatched)`.
- The `state.WebSessionID` branch (updating a legacy interaction UI session) is removed. Only the interaction engine sets `WebSessionID` on a code, and no Auth UI handler drives it any more: `Controller.EntryPointGet`, `EntryPointPost`, `InteractionGet` and `InteractionPost` have no callers.

`pkg/auth/deps.go`: remove `wire.Bind(new(handlerwebappauthflowv2.AuthenticationFlowWebsocketEventStore), new(*authenticationflow.WebsocketEventStore))`. `handlerwebapp.OTPCodeService` stays: three settings handlers use it.

Regenerate `pkg/auth/wire_gen.go` with `make generate` in each commit that changes wiring.

## 4. Event / delivery flow

Unchanged: `approve` publishes the same `refresh` event; `inspect` publishes nothing.

## 5. Compatibility and deployment behavior

- `POST /api/v1/login_link/approve`: request, response and errors unchanged. Validation runs the same schema call on the same bytes, now inside `APIController`.
- `POST /api/v1/login_link/inspect`: additive.
- Auth UI, on submit only: approving now runs `VerifyOTP`. A link used or replaced by a resend after the page opened shows `invalid_code` instead of `matched`. Today such a submit writes the old code into the new code's record, and the flow's `check` then fails and counts a failed attempt. Page load is unchanged.
- No storage change. Not a breaking change.

## 6. File-level change plan

| File | Change |
|---|---|
| `docs/specs/login-link.md` | `inspect` endpoint, approval page uses it, shared request/response/errors |
| `pkg/lib/feature/loginlink/service.go`, `deps.go` | new: `Service` with `Inspect`, `Approve` |
| `pkg/lib/feature/loginlink/service_test.go`, `service_mock_test.go` | new: tests, generated mock |
| `pkg/lib/deps/deps_common.go` | set and two binds |
| `pkg/auth/api/api_controller.go` (+ `_test.go`, `_mock_test.go`), `pkg/auth/api/deps.go` | new: schema, request, `APIController` |
| `pkg/auth/handler/api/login_link_v1_approve.go` (+ test; its mock file is deleted) | read body, call `ApproveLoginLink` |
| `pkg/auth/handler/api/login_link_v1_inspect.go` (+ test) | new handler |
| `pkg/auth/handler/api/deps.go`, `pkg/auth/routes.go`, `pkg/auth/wire_handler.go` | wiring |
| `pkg/auth/handler/webapp/authflowv2/verify_login_link.go` | call `APIController` |
| `pkg/auth/deps.go`, `pkg/auth/wire_gen.go`, `.vettedpositions` | binds, regenerate, vet the new handler's `r.Context()` |
| `e2e/tests/login_link/inspect.test.yaml` | new |

## 7. Test plan

Convey with gomock, as in `login_link_v1_approve_test.go`.

- `loginlink/service_test.go` (mocks for `OTPCodeService` and `EventStore`):
  - `Inspect` valid code: lookup, `VerifyOTP` with `SkipConsume`; no `SetSubmittedCode`, no publish.
  - `Approve` valid code: lookup, `VerifyOTP`, `SetSubmittedCode`, publish `refresh`.
  - `Approve` without a websocket channel: no publish.
  - Both: lookup miss → `InvalidOTPCode`; a `VerifyOTP` error (`ErrConsumedCode`, `RateLimited`) is returned and nothing is submitted.
- `pkg/auth/api/api_controller_test.go` (Convey; mock `APIControllerLoginLinkService`):
  - `map[string]any{"code": "C"}` reaches the service as `"C"`, for `InspectLoginLink` and `ApproveLoginLink`; the service's error is returned.
  - `{}`, `{"code": ""}` and `{"code": "C", "x": 1}` → `ValidationFailed` "invalid request body", and the service is not called.
- `login_link_v1_approve_test.go` and `login_link_v1_inspect_test.go` (a real `authapi.APIController` over a stub `APIControllerLoginLinkService` in the test file, since the generated mock in `pkg/auth/api` is test-only):
  - success → 200 `{"result":{}}`;
  - `InvalidOTPCode` from the service → 403;
  - missing `code` → 400, and the service is not called;
  - a non-JSON `Content-Type` → 400;
  - a malformed body (`{`, `not json`, empty, `[]`) is not passed to the service.
- Auth UI page handler: no unit test, matching the other authflowv2 page handlers.
- E2E, `e2e/tests/login_link/inspect.test.yaml` (`email_otp_mode: login_link`, as `approve.test.yaml`):
  1. inspect a sent link → 200; the flow's `can_check` is still `false`; approve → 200; inspect again → 200 (an approved link can still be approved).
  2. inspect `"code": "invalid"` → 403 `InvalidOTPCode`.
  3. inspect after the flow consumed the code with `check` → 403 `InvalidOTPCode`.
  4. CORS: inspect from a Custom UI origin returns `Access-Control-Allow-Origin` for it.
  5. Auth UI: opening `/authflow/v2/verify_login_link?code=UNKNOWN` redirects to `state=invalid_code`, which exercises the page's call to `APIController`.

## 8. Fixed behavioral decisions

- `loginlink.Service` takes the code. `APIController` owns the request: it validates the body with the endpoint's schema and passes the code on. The HTTP handlers and Auth UI both call `APIController`, so Auth UI can do only what the API allows.
- `APIController` lives in `pkg/auth/api`, the API-side counterpart of `pkg/auth/webapp`. Not in `AuthflowController` or `pkg/auth/webapp` (the API would depend on Auth UI), and not in `pkg/auth/handler/api` (Auth UI would import a handler package it does not import today).
- Auth UI calls it in process; it does not call the HTTP API from the browser, and does not call the OTP service.
- The endpoint is `inspect`, to avoid "check", which in the Authentication Flow API is the input that completes the step.
- `inspect` and `approve` share one request schema, result and error set.
- The legacy web session refresh is removed as unreachable.
- Latte keeps its own logic.

## 9. Implementation order

Spec → service → controller and approve handler → inspect endpoint → Auth UI → e2e.

## 10. Atomic commit plan

1. `doc: Add POST /api/v1/login_link/inspect to the login link spec` — `docs/specs/login-link.md`.
2. `doc: Add plan for Auth UI login link approval through the API` — this file.
3. `Add loginlink.Service for login link approval`
   - `pkg/lib/feature/loginlink/*`, `pkg/lib/deps/deps_common.go`. The set is not used yet, so `wire_gen.go` should not change; regenerate to confirm.
4. `Approve login links through APIController`
   - `pkg/auth/api/api_controller.go` (+ test, mock), `pkg/auth/api/deps.go`, `login_link_v1_approve.go` (+ test; delete `login_link_v1_approve_mock_test.go`), `pkg/auth/handler/api/deps.go`, `pkg/auth/deps.go`, `pkg/auth/wire_gen.go`. Update `.vettedpositions` with `make update-vettedpositions` if the handler's `r.Context()` moved.
5. `Add POST /api/v1/login_link/inspect`
   - `login_link_v1_inspect.go` (+ test), `pkg/auth/handler/api/deps.go`, `pkg/auth/routes.go`, `pkg/auth/wire_handler.go`, `pkg/auth/wire_gen.go`, `.vettedpositions` (new `r.Context()`).
6. `Approve login links in Auth UI through APIController`
   - `verify_login_link.go`, `pkg/auth/deps.go`, `pkg/auth/wire_gen.go`.
7. `Add login link inspect e2e tests` — `e2e/tests/login_link/inspect.test.yaml`.
