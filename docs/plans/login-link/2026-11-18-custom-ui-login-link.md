# Login link approval API and Custom UI endpoints — implementation plan

Specs: [login-link.md](../../specs/login-link.md), [custom-ui.md](../../specs/custom-ui.md).

## 1. Goal / scope

In scope:

- `ui.custom_ui_endpoints` (project) config. A static client's `x_custom_ui_uri` overrides the `authentication` page.
- Resolution of the `authentication`, `login_link` and `account_recovery` pages, used by:
  - the authorize endpoint (where the end-user is sent, HTTP 200 vs 303, implicit redirect URI allowance);
  - the OTP message sender (where login links and account recovery links point).
- Origins: CORS, `frame-ancestors`, passkey RP origins.
- `POST /api/v1/login_link/approve`.
- Default login-link templates use `{{ .Link }}` instead of `{{ .URL }}`.
- A warning log when a redirect URI is allowed only by the Custom UI origin rule, to find out who relies on that rule before removing it.

Out of scope:

- Per-client `login_link` and `account_recovery` pages. Per-client overrides were judged confusing; only `x_custom_ui_uri` stays per client.
- Portal UI for editing `ui.custom_ui_endpoints`, to be designed later. The portal keeps editing `x_custom_ui_uri`.
- SAML (`samlsession.UIService.BuildAuthenticationURL` keeps using Auth UI).
- Legacy interaction/latte UIs: they call the same `otp.MessageSender`, so they pick up the new link resolution automatically. Nothing else changes for them.
- `{"check": true}` before approval counting as a failed OTP attempt (`node_authn_oob.go` `IsCheck` branch). Known, not changed here.

## 2. Config model and schema

### New file `pkg/lib/config/custom_ui.go`

```go
var _ = Schema.Add("CustomUIEndpointsConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"authentication": { "type": "string", "format": "uri" },
		"login_link": { "type": "string", "format": "uri" },
		"account_recovery": { "type": "string", "format": "uri" }
	}
}
`)

type CustomUIEndpointsConfig struct {
	Authentication  string `json:"authentication,omitempty"`
	LoginLink       string `json:"login_link,omitempty"`
	AccountRecovery string `json:"account_recovery,omitempty"`
}
```

`format: uri` is the same check `x_custom_ui_uri` uses today, and it requires an absolute URI.

Resolution helpers, in the same file. All are nil-safe on `c` and `client`, because tests may pass a nil `*UIConfig` and there may be no client.

```go
// The client's own authentication page (x_custom_ui_uri), ignoring the project's.
func (c *OAuthClientConfig) CustomUIAuthenticationURI() string

// Each returns "" when Auth UI should be used.
func (c *UIConfig) ResolveCustomUIAuthenticationURI(client *OAuthClientConfig) string
func (c *UIConfig) ResolveCustomUILoginLinkURI() string
func (c *UIConfig) ResolveCustomUIAccountRecoveryURI() string

// Every URI in ui.custom_ui_endpoints, and every client's x_custom_ui_uri. For CORS.
func CustomUIAllURIs(ui *UIConfig, oauth *OAuthConfig) []string
// The project's authentication, and every client's x_custom_ui_uri. For passkey and frame-ancestors.
func CustomUIAuthenticationURIs(ui *UIConfig, oauth *OAuthConfig) []string
```

Resolution order (spec § Resolution):

| Helper | 1 | 2 |
|---|---|---|
| `ResolveCustomUIAuthenticationURI` | `client.CustomUIURI` | `c.CustomUIEndpoints.Authentication` |
| `ResolveCustomUILoginLinkURI` | `c.CustomUIEndpoints.LoginLink` | — |
| `ResolveCustomUIAccountRecoveryURI` | `c.CustomUIEndpoints.AccountRecovery` | — |

`client == nil` skips step 1. The two list helpers skip empty strings. They do not deduplicate, because the consumers already deduplicate by origin.

### `pkg/lib/config/ui.go`

- `UIConfig` gets `CustomUIEndpoints *CustomUIEndpointsConfig \`json:"custom_ui_endpoints,omitempty"\``.
- The `UIConfig` schema gets `"custom_ui_endpoints": { "$ref": "#/$defs/CustomUIEndpointsConfig" }`.

### `pkg/lib/config/oauth.go`

- `UseHTTP200()` becomes `UseHTTP200(ui *UIConfig) bool { return ui.ResolveCustomUIAuthenticationURI(c) != "" }`. That HTTP 200 HTML redirect exists for Custom UIs (commit b39f2eeeec), so it follows the resolved page, including the project's.
- `x_custom_ui_uri` is unchanged: it stays the client's `authentication` page.

### Feature gate: `pkg/lib/config/configsource/resources.go`

The existing `!*fc.OAuth.Client.CustomUIEnabled` block, which already rejects `x_custom_ui_uri`, also rejects any non-empty key of `appConfig.UI.CustomUIEndpoints`, at `/ui/custom_ui_endpoints` with `custom ui is disallowed`. No feature-config change is needed.

### Migration

None. Old configs keep their meaning. Nothing is rewritten.

## 3. Runtime flow

### 3.1 Authentication page (authorize endpoint)

- `pkg/lib/oauth/oidc/ui.go`
  - `UIURLBuilder` gets `UIConfig *config.UIConfig`.
  - `BuildAuthenticationURL`: `customUIURI := b.UIConfig.ResolveCustomUIAuthenticationURI(client)`, and `isCustomUI := customUIURI != ""`. The rest is unchanged: `BuildCustomUIEndpoint(customUIURI)`, plus forwarding `login_hint`/`id_token_hint`.
- `pkg/lib/oauth/handler/resolve.go`
  - The "same origin as custom ui uri" block uses `client.CustomUIAuthenticationURI()`, the client's own `x_custom_ui_uri`. The project's `authentication` page is not used, so no client, including DCR/CIMD clients, gains a redirect URI it did not configure.
- `pkg/lib/oauth/handler/handler_authz.go`
  - `AuthorizationHandler` gets `UIConfig *config.UIConfig`.
  - `client.UseHTTP200()` becomes `client.UseHTTP200(h.UIConfig)` (lines 764, 833) and `opts.Client.UseHTTP200(h.UIConfig)` (line 915).
- `pkg/lib/oauth/handler/handler_proxy_redirect.go`
  - `ProxyRedirectHandler` gets `UIConfig *config.UIConfig`.
  - Pass `h.UIConfig` to `UseHTTP200`.
- `pkg/lib/oauth/oidc/handler/handler_end_session.go`
  - `EndSessionHandler` gets `UIConfig *config.UIConfig`, used in `UseHTTP200`.

`*config.UIConfig` is already provided by `wire.FieldsOf(new(*config.AppConfig), "UI")` (`pkg/lib/deps/deps_config.go`). The `wire_gen.go` files are regenerated.

### 3.2 Origins

- `pkg/lib/infra/middleware/cors_matcher.go`
  - `CORSMatcher` gets `UIConfig *config.UIConfig`.
  - In the per-client loop, replace `append(..., oauthClient.CustomUIURI)` with one `allowedOrigins = append(allowedOrigins, config.CustomUIAllURIs(m.UIConfig, m.OAuthConfig)...)` outside the loop.
  - The final `slice.Deduplicate` stays.
- `pkg/auth/webapp/dynamic_csp_middleware.go`
  - `DynamicCSPMiddleware` gets `UIConfig *config.UIConfig`.
  - The `AllowFrameAncestorsFromCustomUI` branch loops over `config.CustomUIAuthenticationURIs(m.UIConfig, m.OAuthConfig)` and adds `urlutil.ExtractOrigin(u)` for each.
- `pkg/lib/feature/passkey/config_service.go`
  - `ConfigService` gets `UIConfig *config.UIConfig`.
  - `makeRPOrigins` loops over `config.CustomUIAuthenticationURIs(s.UIConfig, s.OAuthConfig)` instead of the per-client `CustomUIURI`. It keeps the existing parse-skip and `seen` dedup.
  - Its early `if s.OAuthConfig == nil` return stays.

### 3.3 Link pages (OTP sender)

`pkg/lib/authn/otp/sender.go`:

```go
type MessageSender struct {
	... existing ...
	UIConfig *config.UIConfig
}

func (s *MessageSender) buildLinkURL(ctx context.Context, msgType translation.MessageType, opts SendOptions) *neturl.URL
```

- `setupTemplateContext` gets `ctx context.Context` as its first parameter, for logging. Callers `sendEmail` and `sendSMS` pass it.
- `buildLinkURL` (replaces the inline switch in `setupTemplateContext`):
  1. Login-link message types (`SetupPrimaryOOB`, `SetupSecondaryOOB`, `AuthenticatePrimaryOOB`, `AuthenticateSecondaryOOB`):
     - `customURI := s.UIConfig.ResolveCustomUILoginLinkURI()`.
     - If it is set, `u, _ := neturl.Parse(customURI)` (already validated by the schema; on a parse error, fall back to Auth UI and log a warning).
     - Otherwise `u = s.Endpoints.LoginLinkVerificationEndpointURL()`.
  2. `MessageTypeForgotPassword`: the same, with `ResolveCustomUIAccountRecoveryURI` and `s.Endpoints.ResetPasswordEndpointURL()`.
     - `x_from_admin_api=true` is set only on the Auth UI URL, since it is unused and not part of the Custom UI contract.
  3. `q := u.Query(); q.Set("code", opts.OTP); u.RawQuery = q.Encode()`. Existing query parameters of the Custom UI URI are kept.
  4. Any other type panics, as today.
- Template variables: `URL` and `Link` are both set to `buildLinkURL(...).String()`, as today. `URL` stays for customized templates.

`otp.DependencySet` uses `wire.Struct(new(MessageSender), "*")`, and `*config.UIConfig` is already provided. Regenerate every `wire_gen.go` that builds `MessageSender`. `make generate` covers them.

### 3.4 `POST /api/v1/login_link/approve`

New file `pkg/auth/handler/api/login_link_v1_approve.go`:

```go
func ConfigureLoginLinkV1ApproveRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "POST").WithPathPattern("/api/v1/login_link/approve")
}

var LoginLinkV1ApproveRequestSchema = validation.NewSimpleSchema(`{
	"type": "object",
	"additionalProperties": false,
	"properties": { "code": { "type": "string", "minLength": 1 } },
	"required": ["code"]
}`)

type LoginLinkV1ApproveRequest struct { Code string `json:"code"` }

type LoginLinkV1ApproveOTPCodeService interface {
	LookupCode(ctx context.Context, purpose otp.Purpose, code string) (target string, err error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error)
}

type LoginLinkV1ApproveEventStore interface {
	Publish(ctx context.Context, websocketChannelName string, e authflow.Event) error
}

type LoginLinkV1ApproveHandler struct {
	Config   *config.AppConfig
	OTPCodes LoginLinkV1ApproveOTPCodeService
	Events   LoginLinkV1ApproveEventStore
}

func (h *LoginLinkV1ApproveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (h *LoginLinkV1ApproveHandler) approve(ctx context.Context, code string) error
```

Call flow:

1. `ServeHTTP`: `httputil.BindJSONBody(r, w, LoginLinkV1ApproveRequestSchema.Validator(), &req)`. On error, write `api.Response{Error: err}`.
2. `approve`:
   1. `kind := otp.KindOOBOTPLink(h.Config, model.AuthenticatorOOBChannelEmail)`.
   2. `target, err := h.OTPCodes.LookupCode(ctx, kind.Purpose(), code)`. An unknown or expired code yields `InvalidOTPCode` (403).
   3. `h.OTPCodes.VerifyOTP(ctx, kind, target, code, &otp.VerifyOptions{SkipConsume: true})`. This applies the validate rate limits (`RateLimited`, 429) and the failed-attempt revocation. It returns `ErrConsumedCode` (`InvalidOTPCode`, 403) for a consumed code, and it does not consume.
   4. `state, err := h.OTPCodes.SetSubmittedCode(ctx, kind, target, code)`.
   5. If `state.AuthenticationFlowWebsocketChannelName != ""`, call `h.Events.Publish(ctx, channel, authflow.NewEventRefresh())`.
   6. Return nil.
3. On success: `api.Response{Result: struct{}{}}`, which serializes to `{"result":{}}`. On error: `api.Response{Error: err}`. A non-API error is logged at error level, as `AnonymousUserPromotionCodeAPIHandler` does.

Approving twice repeats steps 2.2–2.5 and succeeds, as the spec requires. Unlike the Auth UI handler, there is no web-session update: `state.WebSessionID` is only set by the legacy interaction UI, which never uses this API.

Wiring:

- `pkg/auth/routes.go`: `router.Add(apihandler.ConfigureLoginLinkV1ApproveRoute(authenticationFlowRoute), p.Handler(newAPILoginLinkV1ApproveHandler))`. `authenticationFlowRoute` gives the same CORS, intl and rate-limit middleware as the Authentication Flow API.
- `pkg/auth/wire_handler.go`: `newAPILoginLinkV1ApproveHandler`, following `newAPIAuthenticationFlowV1GetHandler`.
- `pkg/auth/deps.go`: add the two bindings:
  - `wire.Bind(new(handlerapi.LoginLinkV1ApproveOTPCodeService), new(*otp.Service))`
  - `wire.Bind(new(handlerapi.LoginLinkV1ApproveEventStore), new(*authenticationflow.WebsocketEventStore))`
- Regenerate `pkg/auth/wire_gen.go`.

### 3.5 Redirect URI allowed only by the Custom UI origin

`validateRedirectURI` (`pkg/lib/oauth/handler/resolve.go`) implicitly allows a redirect URI on the origin of the client's `x_custom_ui_uri`. Nothing documents this rule, and `authgear-nwd-ui` may rely on it through `GET /oauth2/redirect`. To find out who does before removing it:

- `validateRedirectURI(...) (onlyCustomUIOrigin bool, err error)` reports when that rule is the only one allowing the URI.
- `logCustomUIOriginRedirectURI(ctx, clientIDs, redirectURI)` logs a warning, with the client IDs and the redirect URI, on the `oauth-redirect-uri` logger.
- Callers:
  - `parseRedirectURI` gets `ctx` and logs, which covers the authorize endpoint and app2app at the token endpoint;
  - the authorize error redirect (`prepareConsentErrInvalidOAuthResponse`) logs;
  - `ProxyRedirectHandler.Validate` gets `ctx`, and logs only when no client allows the URI through another rule, since it checks every client.

## 4. Event / delivery flow

Only the link inside the message changes; delivery is unchanged. The websocket `refresh` event is published on the channel stored on the code (`Code.AuthenticationFlowWebsocketChannelName`), as the Auth UI handler does.

## 5. Compatibility and deployment behavior

- Config: additive. `x_custom_ui_uri` behaves exactly as before when no new key is set:
  - redirect to the Custom UI;
  - HTTP 200 redirect;
  - implicit redirect URI origin;
  - CORS, CSP and passkey origins.
- New project-level `authentication` behavior applies only when it is set:
  - every client, including DCR/CIMD, is sent there and gets HTTP 200 redirects.
- Emails: with no new key set, links are byte-identical to today, because `buildLinkURL` builds the same URL and the template rename keeps the rendered output.
- Storage: no Redis key or payload change. The approve API uses the existing OTP code and lookup stores.
- Templates: projects with customized templates that use `{{ .URL }}` keep working, because `URL` is still set.
- Not a breaking change, so there is no `BREAKING-CHANGES.md` entry.
- Rollout: an old pod rejects a config containing the new keys (`additionalProperties: false`). Set the new keys only after every pod runs the new version. Mention this in the PR description.

## 6. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/config/custom_ui.go` | new: struct, schema, resolution helpers |
| `pkg/lib/config/custom_ui_test.go` | new: resolution tests |
| `pkg/lib/config/ui.go` | `CustomUIEndpoints` field and schema |
| `pkg/lib/config/oauth.go` | `UseHTTP200(ui)` |
| `pkg/lib/config/testdata/config_tests.yaml` | parse cases |
| `pkg/lib/config/configsource/resources.go` (+ `_test.go`) | feature gate |
| `pkg/lib/oauth/oidc/ui.go` (+ `ui_test.go`) | resolved authentication page |
| `pkg/lib/oauth/handler/resolve.go`, `handler_authz.go`, `handler_proxy_redirect.go` (+ `_test.go`) | `UIConfig` plumbing |
| `pkg/lib/oauth/oidc/handler/handler_end_session.go` | `UIConfig` plumbing |
| `pkg/lib/infra/middleware/cors_matcher.go` (+ `cors_test.go`) | origins |
| `pkg/auth/webapp/dynamic_csp_middleware.go` (+ test) | origins |
| `pkg/lib/feature/passkey/config_service.go` (+ test) | origins |
| `pkg/lib/authn/otp/sender.go`, new `sender_test.go` | link resolution |
| `pkg/auth/handler/api/login_link_v1_approve.go` (+ `_test.go`, mock) | approve API |
| `pkg/auth/routes.go`, `wire_handler.go`, `deps.go` | wiring |
| `pkg/lib/oauth/handler/resolve.go`, `handler_authz.go`, `handler_token.go`, `handler_proxy_redirect.go`, `pkg/auth/handler/oauth/redirect.go` | redirect URI log |
| `resources/authgear/templates/**/*_login_link*`, `messages/translation.json` | `URL` → `Link` |
| `e2e/tests/login_link/**`, `e2e/tests/custom_ui_endpoints/**`, `e2e/cmd/smtp`, `e2e/pkg/testrunner` | e2e |
| generated: `wire_gen.go` files | regenerate |

## 7. Test plan

All unit tests use Convey, matching the packages touched.

- `custom_ui_test.go`:
  - `x_custom_ui_uri` overrides the project's `authentication` page only;
  - project fallback;
  - nil client, nil `CustomUIEndpoints`, nil `UIConfig` → project or "";
  - the list helpers include and exclude the right URIs.
- `config_tests.yaml`:
  - valid `ui.custom_ui_endpoints`;
  - a relative URI and a non-URI are rejected (format uri).
  - Error strings are captured from real `config.Parse` output.
- `resources_test.go`: with the feature disabled, `ui.custom_ui_endpoints` is rejected.
- `ui_test.go`: project-level authentication is used when the client has none; `x_custom_ui_uri` wins; Auth UI when nothing is set.
- `resolve_test.go`: a redirect URI on the project page's origin is rejected; `validateRedirectURI` reports when only the Custom UI origin allows the URI.
- `handler_proxy_redirect_test.go`: HTTP 200 when only the project-level authentication is set.
- `cors_test.go`: the origins of the project pages and of a client's `x_custom_ui_uri` are allowed.
- `dynamic_csp_middleware_test.go`: the project authentication origin is in frame-ancestors; login_link and account_recovery are not.
- `passkey` test: the project authentication origin is in the RP origins; login_link and account_recovery are not.
- `sender_test.go` (mock `EndpointsProvider`; assert on the built URL and `PartialTemplateVariables`):
  - Auth UI login link (unchanged URL);
  - Custom UI login link with an existing query `?a=b` → `?a=b&code=...`;
  - forgot password with Admin API → Auth UI URL with `x_from_admin_api`;
  - forgot password with Custom UI → no `x_from_admin_api`.
- `login_link_v1_approve_test.go` (mocks):
  - success publishes refresh;
  - no channel → no publish;
  - lookup miss → 403;
  - `VerifyOTP` error is propagated (403 / 429);
  - missing `code` → 400.
- E2E `e2e/tests/login_link/` (YAML; `email_otp_mode: login_link`):
  1. login: identify by email → `authenticate` returns `otp_form: link` → `http_request` POST `/api/v1/login_link/approve` with `{{ linkOTPCode "email" ... }}` → 200 `{"result":{}}` → input `{"check": true}` → flow finishes.
  2. approve the same code twice → both 200.
  3. approve `"code": "invalid"` → 403 `InvalidOTPCode`.
  4. approve after the flow consumed the code → 403 `InvalidOTPCode`.
  - Signup `create_authenticator` is not covered: `linkOTPCode` derives the code from an existing user, and the user does not exist until signup finishes.
- E2E `e2e/tests/custom_ui_endpoints/`. The e2e SMTP server records the decoded text body of each mail, and `smtp_log_query` takes an optional `smtp_log_text_contains`, so a test can assert the link in an email:
  1. authorize goes to the project `authentication` page, keeping its query parameters;
  2. a client's `x_custom_ui_uri` beats the project page;
  3. a redirect URI on the project page's origin is rejected;
  4. CORS allows the origins of a project `login_link` page and a client's `x_custom_ui_uri`, and no other;
  5. a login link opens the project `login_link` page (query parameters kept), and approving from that origin then `check` finishes the flow;
  6. a login link opens Auth UI when only `authentication` is set;
  7. an account recovery link opens the project `account_recovery` page, and its code continues the flow.

## 8. Fixed behavioral decisions

- `UseHTTP200` follows the client's resolved `authentication` page, including the project default. The implicit redirect-URI origin follows only the client's own `x_custom_ui_uri`.
- Login link and account recovery links go to the project's pages, even for a client that sets `x_custom_ui_uri`.
- The new key is gated by the existing `oauth.client.custom_ui_enabled` feature flag.
- `x_from_admin_api` is only added to Auth UI account recovery links.
- The approve API uses `VerifyOTP` with `SkipConsume`, so it enforces rate limits and failed-attempt revocation, which is stricter than the Auth UI handler. The Auth UI handler is not changed.
- An unparseable resolved URI (impossible after schema validation) falls back to Auth UI with a warning log rather than failing the send.

## 9. Implementation order

Config → authentication page → origins → sender → approve API → templates → e2e → redirect URI log.

## 10. Atomic commit plan

1. `doc: Add login link and Custom UI endpoints implementation plan` — this file.
2. `Add Custom UI endpoints config`
   - Files: `custom_ui.go` + test, `ui.go`, `oauth.go`, `config_tests.yaml`, `configsource/resources.go` + test.
   - Changes `UseHTTP200` to take `ui` and updates its 5 call sites by threading `UIConfig` through `AuthorizationHandler`, `ProxyRedirectHandler` and `EndSessionHandler` so the commit compiles.
   - Includes `oidc/ui.go` resolution, the `resolve.go` redirect URI rule, and their tests.
   - Regenerate `wire_gen.go` in the same commit.
3. `Resolve Custom UI origins from custom_ui_endpoints`
   - Files: CORS matcher, CSP middleware, passkey config service, and their tests.
   - Regenerate wire.
4. `Point login and account recovery links to the Custom UI`
   - Files: `sender.go`, `sender_test.go`.
   - Regenerate wire.
5. `Add POST /api/v1/login_link/approve`
   - Files: handler, test, mock, `routes.go`, `wire_handler.go`, `deps.go`, and its `.vettedpositions` entry.
   - Regenerate wire.
6. `Use Link instead of URL in default login link templates`
   - Follows the `update-email-templates` skill: the mjml sources, the `translation.json` values in every locale, and the regenerated outputs, in separate commits.
7. `Add login link approve e2e tests` — `e2e/tests/login_link/`, and the regenerated e2e wiring.
8. `Add Custom UI endpoints e2e tests` — the SMTP mail log, `smtp_log_text_contains`, and `e2e/tests/custom_ui_endpoints/`.
9. `Log redirect URIs allowed only by the Custom UI origin` — §3.5, and its `.vettedpositions` entry.
