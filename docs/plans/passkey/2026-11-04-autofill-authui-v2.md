# Passkey AutoFill in AuthUI v2

Spec: [docs/specs/webauthn.md](../../specs/webauthn.md#autofill-for-passkeys).

## 1. Goal and scope

Offer passkey AutoFill (WebAuthn conditional mediation) on the login ID input of:

- `authflowv2/login.html`;
- `authflowv2/signup.html`, in the `signup_login` UI variant only (`ui.signup_login_flow_enabled: true`).

Both email/username (text) and phone inputs are covered. Choosing a passkey posts `x_action=passkey`, which both pages already handle (`AuthflowV2LoginHandler` in `pkg/auth/handler/webapp/authflowv2/login.go`, `InternalAuthflowV2SignupLoginHandler` in `internal_signup_login.go`).

No server, config or schema change. Unchanged and out of scope: `/_internals/passkey/request_options`, `/_internals/passkey/creation_options`, `PasskeyRequestController`, `PasskeyCreationController`, passkey error matching (`handleError`, `AuthflowV2PasskeyErrorController`).

## 2. Challenge source

`PasskeyAutofillController` today fetches options from `POST /_internals/passkey/request_options?conditional=true` on every attempt. It stops doing that and reads the options the page already has, the same `$.PasskeyRequestOptionsJSON` the "Continue with passkey" button uses.

Why that is safe:

- `NewIntentLoginFlowStepIdentify` and `NewIntentSignupLoginFlowStepIdentify` call `PasskeyRequestOptionsService.MakeModalRequestOptions` once, when the step is created, and keep the result in flow state. The page renders the same options until the step is left.
- Verifying an assertion only peeks the challenge (`passkey.Service.PeekAssertionResponse`). It is consumed by `ConsumeAssertionResponse` in an `OnCommitEffect` of `NodeDoUseIdentityPasskey`, which runs only when the flow finishes. A rejected passkey leaves the challenge valid, so retrying with the same options works.
- The options have no `allowCredentials` (`AllowCredentials: nil`, omitted from JSON), as conditional mediation requires, and `timeout` is `MediationConditionalTimeout` (1 hour), longer than the flow.

The browser adds `mediation: "conditional"` to the deserialized options. `deserializeRequestOptions` mutates its argument, so each attempt parses the JSON again.

## 3. Runtime flow (`authui/src/passkey.ts`)

### 3.1 `PasskeyAutofillController`

- `static values = { options: String }`, `declare optionsValue: string`.
- `connect()`: listen to `PASSKEY_MODAL_START` and `PASSKEY_MODAL_END` on `document`, then `start()`.
- `disconnect()`: remove both listeners, then `stop()`.
- `start()`: if `passkeyIsAvailable()`, create the run's `AbortController` and `RetryEventTarget` (retry calls `setupAutofill()`), then `setupAutofill()`.
- `stop()`: abort the run's `AbortController` (which also cancels a scheduled retry) and clear both fields.
- `onModalStart = () => this.stop()`; `onModalEnd = () => { this.stop(); this.start(); }`.
- `setupAutofill()`:
  1. Capture `this.abortController` and `this.retryEventTarget`; return if either is null. A run stopped while awaiting must neither call `navigator.credentials.get` nor schedule a retry, so every later step checks the captured `signal.aborted`, never `this.*`.
  2. Return if `PublicKeyCredential.isConditionalMediationAvailable` is missing or resolves to `false`, or if the run was stopped meanwhile.
  3. `options = deserializeRequestOptions(JSON.parse(this.optionsValue))`, then `navigator.credentials.get({ ...options, mediation: "conditional", signal })`.
  4. On a credential: serialize it into the `input` target and click the `submit` target, then schedule a retry, as today. The next page render disconnects the controller and aborts it.
  5. On an error: schedule a retry, unless the run was stopped or the error is an `AbortError`. Reset the backoff (`markSuccess()`) first when the request stayed pending for at least 500 ms, so a user who cancels gets suggestions back after 1 s, while a request that fails at once (the iOS 16 beta bug) backs off 1, 2, 4 s.
  6. Errors go to the existing `handleError`, unchanged.

### 3.2 `AuthflowPasskeyRequestController._use()`

A page can have only one pending `navigator.credentials.get`; a modal request made while a conditional one is pending is rejected at once with `NotAllowedError`.

- Before the modal request, `document.dispatchEvent(new CustomEvent(PASSKEY_MODAL_START))`. Listeners run synchronously, so the conditional request is aborted before `navigator.credentials.get`, still inside the click's user activation.
- In `finally`, dispatch `PASSKEY_MODAL_END`, whether the request succeeded or not. After a success the form is being submitted; the next page render disconnects the controller and aborts the restarted request.

Constants, next to the controller: `PASSKEY_MODAL_START = "passkey:modal-start"`, `PASSKEY_MODAL_END = "passkey:modal-end"`.

## 4. Templates

### 4.1 Registration

`authui/src/authflowv2.ts`: `Stimulus.register("passkey-autofill", PasskeyAutofillController)`.

### 4.2 Hidden form

After the login ID `<form>` (forms cannot nest), on both pages:

```html
<div class="hidden"
  data-controller="passkey-autofill"
  data-passkey-autofill-options-value="{{ $.PasskeyRequestOptionsJSON }}">
  <form method="post" novalidate data-controller="turbo-form" data-action="submit->turbo-form#submitForm">
    <input type="hidden" name="x_assertion_response" data-passkey-autofill-target="input">
    <button type="submit" class="hidden" name="x_action" value="passkey" data-passkey-autofill-target="submit"></button>
  </form>
</div>
```

Rendered when the page shows the passkey button and a login ID input, and is not an inline preview (whose options are `{}`):

- `login.html`: `and $show_input_form $.PasskeyEnabled (not $.AlternativesDisabled) (not $.InlinePreview)`, inside `{{ if $has_login_id }}`.
- `signup.html`: the same, plus `eq $.UIVariant "signup_login"`.

`$.AlternativesDisabled` is set by an enforced login hint, which also hides the button; AutoFill must not let the user sign in as someone else there.

### 4.3 Phone input

`authflowv2/__phone_input.html` hardcodes `autocomplete="username"` on the visible phone input. It takes a new `Autocomplete` parameter, defaulting to `username`, so every other caller is unchanged:

```
autocomplete="{{ or $.Autocomplete "username" }}"
```

- `login.html` passes `"Autocomplete" "username webauthn"`, matching its text input.
- `signup.html` passes `username webauthn` in the `signup_login` variant and `username` otherwise; its text input does the same. The plain sign-up page renders as before.

`.make-lint-translation-keys-expect` is regenerated in each commit that moves template lines.

## 5. Compatibility

- No stored data, config or API change. Turning AutoFill off needs no migration: removing the hidden form restores today's behaviour.
- Custom templates that override `login.html` or `signup.html` get no AutoFill, and keep working.
- v1 AuthUI registers `PasskeyAutofillController` too, but no v1 template uses it.

## 6. Tests

- `authui/src/passkey.test.ts` (jest, jsdom, Stimulus `Application`, like `authflowv2/clockSkew.test.ts`), with `PublicKeyCredential` and `navigator.credentials.get` stubbed:
  - the request uses the page's options, with `mediation: "conditional"` and the decoded challenge, and makes no HTTP request;
  - a returned credential fills `x_assertion_response` and clicks the submit button;
  - after a rejection, it retries with the same options;
  - a request that fails at once backs off; one that was pending resets the backoff;
  - `passkey:modal-start` aborts the pending request and `passkey:modal-end` starts a new one;
  - a run stopped while `isConditionalMediationAvailable` is pending never calls `navigator.credentials.get`;
  - no request when conditional mediation is unavailable.
- `pkg/auth/handler/webapp/authflowv2/passkey_autofill_test.go` (Convey, rendering the real templates like `clock_skew_test.go`):
  - `login.html` renders the hidden form with the options, for text and phone input, and the phone input has `username webauthn`;
  - no hidden form without passkey, with `AlternativesDisabled`, or in inline preview;
  - `signup.html` in `signup_login` renders the hidden form and `username webauthn` on text and phone input;
  - `signup.html` in the `signup` variant renders neither, and its inputs keep `autocomplete="username"`.
- e2e: none. The e2e runner drives authflow over the API and has no browser, so it cannot exercise conditional mediation. The server path (`x_action=passkey` on these pages) is unchanged.
- Manual: Chrome, Safari and Firefox on desktop, Safari on iOS, Chrome on Android; email, username and phone inputs; both pages; the button while AutoFill is pending; cancelling and seeing suggestions come back.

## 7. Atomic commits

1. `doc: Specify passkey AutoFill on the AuthUI login pages`: `docs/specs/webauthn.md`.
2. `doc: Add passkey AutoFill implementation plan`: this file.
3. `Take passkey autofill options from the page`: `authui/src/passkey.ts` (`PasskeyAutofillController`: options value, `mediation: "conditional"`, `start()`/`stop()`, stale-run guard, backoff reset), `authui/src/passkey.test.ts`.
4. `Pause passkey autofill while the passkey dialog is open`: `authui/src/passkey.ts` (events in `AuthflowPasskeyRequestController` and listeners in `PasskeyAutofillController`), `authui/src/passkey.test.ts`.
5. `Offer passkey autofill on the login and signup login pages`: `authui/src/authflowv2.ts`, `login.html`, `signup.html` (hidden form, text input token), `.make-lint-translation-keys-expect`, `passkey_autofill_test.go`.
6. `Offer passkey autofill on the phone login ID input`: `__phone_input.html`, `login.html`, `signup.html`, `.make-lint-translation-keys-expect`, `passkey_autofill_test.go`.
