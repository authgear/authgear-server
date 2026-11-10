import { Controller } from "@hotwired/stimulus";
import axios from "axios";
import { handleAxiosError, showErrorMessage } from "./messageBar";
import { base64DecToArr, base64EncArr } from "./base64";
import { base64URLToBase64, trimNewline, base64ToBase64URL } from "./base64url";
import { setErrorMessage } from "./authflowv2/alert-message";
import {
  getAuthflowV2PasskeyErrorMessageID,
  isCancelOrTimeout,
  isDuplicateError,
  isEmptyAllowCredentialsError,
  isFirefoxSecurityKeyError,
  isOperationFailError,
} from "./passkeyError";

function passkeyIsAvailable(): boolean {
  return (
    typeof window.PublicKeyCredential !== "undefined" &&
    typeof window.navigator.credentials !== "undefined"
  );
}

function deserializeCreationOptions(
  creationOptions: any
): CredentialCreationOptions {
  const base64URLChallenge = creationOptions.publicKey.challenge;
  const challenge = base64DecToArr(base64URLToBase64(base64URLChallenge));
  creationOptions.publicKey.challenge = challenge;

  const base64URLUserID = creationOptions.publicKey.user.id;
  const userID = base64DecToArr(base64URLToBase64(base64URLUserID));
  creationOptions.publicKey.user.id = userID;

  if (creationOptions.publicKey.excludeCredentials != null) {
    for (const c of creationOptions.publicKey.excludeCredentials) {
      c.id = base64DecToArr(base64URLToBase64(c.id));
    }
  }
  return creationOptions;
}

function deserializeRequestOptions(
  requestOptions: any
): CredentialRequestOptions {
  const base64URLChallenge = requestOptions.publicKey.challenge;
  const challenge = base64DecToArr(base64URLToBase64(base64URLChallenge));
  requestOptions.publicKey.challenge = challenge;
  if (requestOptions.publicKey.allowCredentials) {
    for (const c of requestOptions.publicKey.allowCredentials) {
      c.id = base64DecToArr(base64URLToBase64(c.id));
    }
  }
  return requestOptions;
}

// At runtime, we do not assume the return value of
// window.navigator.credentials.get is instanceof PublicKeyCredential.
// It is observed that on Firefox 133 (as of 2024-12-18), window.navigator.credentials.get
// can return something of type Object, when 1Password extension is installed.
//
// Maybe this behavior is due to the fact that Firefox has no native support
// for passkeys yet.
// I have tried to search the issue list with
// https://bugzilla.mozilla.org/buglist.cgi?product=Core&component=DOM%3A%20Web%20Authentication&resolution=---
// but found no reported issues.
//
// Therefore, we use a type guard to assert the return value of
// window.navigator.credentials.get looks like a PublicKeyCredential,
// without checking the actual type.

interface PublicKeyCredentialWithAttestationResponse {
  id: string;
  type: string;
  response: AuthenticatorAttestationResponse;
  getClientExtensionResults(): AuthenticationExtensionsClientOutputs;
}

function isPublicKeyCredentialWithAttestationResponse(
  credential: unknown
): credential is PublicKeyCredentialWithAttestationResponse {
  // For browsers that behave correctly, just use instanceof.
  if (credential instanceof PublicKeyCredential) {
    return true;
  }

  // Otherwise, we fallback to checking the required properties.

  if (credential == null) {
    return false;
  }
  if (typeof credential !== "object") {
    return false;
  }
  // I know it is better to write Object.prototype.hasOwnProperty.call(credential, "id"),
  // but TypeScript is not smart enough to type-inference it has .id afterwards.
  if (!("id" in credential) || typeof credential.id !== "string") {
    return false;
  }
  if (!("type" in credential) || typeof credential.type !== "string") {
    return false;
  }
  if (
    !("response" in credential) ||
    typeof credential.response !== "object" ||
    credential.response == null
  ) {
    return false;
  }
  // Since getClientExtensionResults is a function, it is probably not a own property on credential.
  // So it is legit (and correct) to use the in operator here.
  if (
    !("getClientExtensionResults" in credential) ||
    typeof credential.getClientExtensionResults !== "function"
  ) {
    return false;
  }
  return true;
}

interface PublicKeyCredentialWithAssertionResponse {
  id: string;
  type: string;
  response: AuthenticatorAssertionResponse;
  getClientExtensionResults(): AuthenticationExtensionsClientOutputs;
}

function isPublicKeyCredentialWithAssertionResponse(
  credential: unknown
): credential is PublicKeyCredentialWithAssertionResponse {
  // For browsers that behave correctly, just use instanceof.
  if (credential instanceof PublicKeyCredential) {
    return true;
  }

  // Otherwise, we fallback to checking the required properties.

  if (credential == null) {
    return false;
  }
  if (typeof credential !== "object") {
    return false;
  }
  // I know it is better to write Object.prototype.hasOwnProperty.call(credential, "id"),
  // but TypeScript is not smart enough to type-inference it has .id afterwards.
  if (!("id" in credential) || typeof credential.id !== "string") {
    return false;
  }
  if (!("type" in credential) || typeof credential.type !== "string") {
    return false;
  }
  if (
    !("response" in credential) ||
    typeof credential.response !== "object" ||
    credential.response == null
  ) {
    return false;
  }
  // Since getClientExtensionResults is a function, it is probably not a own property on credential.
  // So it is legit (and correct) to use the in operator here.
  if (
    !("getClientExtensionResults" in credential) ||
    typeof credential.getClientExtensionResults !== "function"
  ) {
    return false;
  }
  return true;
}

function serializeAttestationResponse(
  credential: PublicKeyCredentialWithAttestationResponse
) {
  const response = credential.response;

  const attestationObject = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.attestationObject)))
  );
  const clientDataJSON = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.clientDataJSON)))
  );

  let transports: string[] = [];
  if (typeof response.getTransports === "function") {
    transports = response.getTransports();
  }

  const clientExtensionResults = credential.getClientExtensionResults();

  return {
    id: credential.id,
    rawId: credential.id,
    type: credential.type,
    response: {
      attestationObject,
      clientDataJSON,
      transports,
    },
    clientExtensionResults,
  };
}

function serializeAssertionResponse(
  credential: PublicKeyCredentialWithAssertionResponse
) {
  const response = credential.response;
  const authenticatorData = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.authenticatorData)))
  );
  const clientDataJSON = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.clientDataJSON)))
  );
  const signature = trimNewline(
    base64ToBase64URL(base64EncArr(new Uint8Array(response.signature)))
  );
  const userHandle =
    response.userHandle == null
      ? undefined
      : trimNewline(
          base64ToBase64URL(base64EncArr(new Uint8Array(response.userHandle)))
        );
  const clientExtensionResults = credential.getClientExtensionResults();
  return {
    id: credential.id,
    rawId: credential.id,
    type: credential.type,
    response: {
      authenticatorData,
      clientDataJSON,
      signature,
      userHandle,
    },
    clientExtensionResults,
  };
}

function handleError(err: unknown) {
  // Cancel and timeout are expected, so do not log them as errors.
  if (isCancelOrTimeout(err)) {
    console.info(err);
    return;
  }
  console.error(err);

  if (isOperationFailError(err) || isFirefoxSecurityKeyError(err)) {
    showErrorMessage("error-message-no-passkey");
    return;
  }

  if (isEmptyAllowCredentialsError(err)) {
    showErrorMessage("error-message-passkey-empty-allow-credentials");
    return;
  }

  if (isDuplicateError(err)) {
    showErrorMessage("error-message-passkey-duplicate");
  }
}

// We want to prompt the modal dialog to create passkey.
// But navigator.credentials.create is user activation-gated API.
// There is no way to check if we have user activation currently.
export class PasskeyCreationController extends Controller {
  static targets = ["button", "submit", "input"];

  declare buttonTarget: HTMLButtonElement;
  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;

  connect() {
    // Disable the button if PublicKeyCredential is unavailable.
    if (!passkeyIsAvailable()) {
      this.buttonTarget.disabled = true;
      return;
    }
  }

  create(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    this._create();
  }

  async _create() {
    try {
      const resp = await axios("/_internals/passkey/creation_options", {
        method: "post",
      });
      const options = deserializeCreationOptions(resp.data.result);
      try {
        const rawResponse = await window.navigator.credentials.create(options);
        if (isPublicKeyCredentialWithAttestationResponse(rawResponse)) {
          const response = serializeAttestationResponse(rawResponse);
          const responseString = JSON.stringify(response);
          this.inputTarget.value = responseString;
          // It seems that we should use form.submit() to submit the form.
          // but form.submit() does NOT trigger submit event,
          // which is essential for XHR form submission to work.
          // Therefore, we emulate form submission here by clicking the submit button.
          this.submitTarget.click();
        }
      } catch (e: unknown) {
        handleError(e);
      }
    } catch (e: unknown) {
      handleAxiosError(e);
    }
  }
}

// We want to prompt the modal dialog to let the user to choose passkey.
// But navigator.credentials.get is user activation-gated API.
// There is no way to check if we have user activation currently.
export class PasskeyRequestController extends Controller {
  static targets = ["button", "submit", "input"];
  static values = {
    auto: String,
    allowCredentials: String,
  };

  declare buttonTarget: HTMLButtonElement;
  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;

  declare autoValue: string;
  declare allowCredentialsValue: string;

  connect() {
    // Disable the button if PublicKeyCredential is unavailable.
    if (!passkeyIsAvailable()) {
      this.buttonTarget.disabled = true;
      return;
    }

    if (this.autoValue === "true") {
      this._use();
    }
  }

  use(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    this._use();
  }

  async _use() {
    try {
      const params = new URLSearchParams();
      params.set("allow_credentials", this.allowCredentialsValue);
      const resp = await axios("/_internals/passkey/request_options", {
        method: "post",
        data: params,
      });
      const options = deserializeRequestOptions(resp.data.result);
      try {
        const rawResponse = await window.navigator.credentials.get(options);
        if (isPublicKeyCredentialWithAssertionResponse(rawResponse)) {
          const response = serializeAssertionResponse(rawResponse);
          const responseString = JSON.stringify(response);
          this.inputTarget.value = responseString;
          // It seems that we should use form.submit() to submit the form.
          // but form.submit() does NOT trigger submit event,
          // which is essential for XHR form submission to work.
          // Therefore, we emulate form submission here by clicking the submit button.
          this.submitTarget.click();
        }
      } catch (e: unknown) {
        handleError(e);
      }
    } catch (e: unknown) {
      handleAxiosError(e);
    }
  }
}

export class AuthflowPasskeyErrorController extends Controller {
  connect(): void {
    document.addEventListener("passkey:error", this.onPasskeyError);
  }

  disconnect(): void {
    document.removeEventListener("passkey:error", this.onPasskeyError);
  }

  onPasskeyError = (e: Event) => {
    handleError((e as CustomEvent).detail);
  };
}

export class AuthflowV2PasskeyErrorController extends Controller {
  connect(): void {
    document.addEventListener("passkey:error", this.onPasskeyError);
  }

  disconnect(): void {
    document.removeEventListener("passkey:error", this.onPasskeyError);
  }

  onPasskeyError = (event: Event) => {
    const err: unknown = (event as CustomEvent).detail;
    const messageID = getAuthflowV2PasskeyErrorMessageID(err);
    if (messageID == null) {
      console.info(err);
      return;
    }
    console.error(err);
    setErrorMessage(messageID);
  };
}

// Dispatched on document by AuthflowPasskeyRequestController around its modal
// navigator.credentials.get, so that PasskeyAutofillController can step aside.
const PASSKEY_MODAL_START = "passkey:modal-start";
const PASSKEY_MODAL_END = "passkey:modal-end";

export class AuthflowPasskeyRequestController extends Controller {
  static targets = ["button", "submit", "input"];
  static values = {
    options: String,
    auto: String,
  };

  declare buttonTarget: HTMLButtonElement;
  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;

  declare optionsValue: string;
  declare autoValue: string;
  declare hasButtonTarget: boolean;

  // True while the request is pending, so that more taps do not start more requests.
  // On iOS, the dialog can take a few seconds to open after the previous one is cancelled,
  // and a second request at that time is rejected.
  requesting = false;

  connect() {
    window.addEventListener("pageshow", this.onPageShow);

    if (this.hasButtonTarget) {
      this.buttonTarget.disabled = false;
      // Disable the button if PublicKeyCredential is unavailable.
      if (!passkeyIsAvailable()) {
        this.buttonTarget.disabled = true;
        return;
      }
    }

    if (this.autoValue === "true") {
      this._use();
    }
  }

  disconnect() {
    window.removeEventListener("pageshow", this.onPageShow);
  }

  // The page can be restored from the back-forward cache, for example when the user goes
  // back after a passkey was chosen. Stimulus does not call connect() again then,
  // so the button would stay disabled.
  onPageShow = (e: PageTransitionEvent) => {
    if (!e.persisted) {
      return;
    }
    this.requesting = false;
    if (this.hasButtonTarget && passkeyIsAvailable()) {
      this.buttonTarget.disabled = false;
    }
  };

  use(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    this._use();
  }

  async _use() {
    if (!passkeyIsAvailable()) {
      document.dispatchEvent(
        new CustomEvent("passkey:error", {
          detail: new Error("passkey is not available"),
        })
      );
      return;
    }
    if (this.requesting) {
      return;
    }
    this.requesting = true;
    if (this.hasButtonTarget) {
      this.buttonTarget.disabled = true;
    }
    // A page can only have one pending navigator.credentials.get.
    // Abort the pending autofill request, otherwise this request would be
    // rejected with NotAllowedError immediately.
    document.dispatchEvent(new CustomEvent(PASSKEY_MODAL_START));
    let submitted = false;
    try {
      const optionsJSON = JSON.parse(this.optionsValue);
      const options = deserializeRequestOptions(optionsJSON);
      const rawResponse = await window.navigator.credentials.get(options);
      if (isPublicKeyCredentialWithAssertionResponse(rawResponse)) {
        const response = serializeAssertionResponse(rawResponse);
        const responseString = JSON.stringify(response);
        this.inputTarget.value = responseString;
        // It seems that we should use form.submit() to submit the form.
        // but form.submit() does NOT trigger submit event,
        // which is essential for XHR form submission to work.
        // Therefore, we emulate form submission here by clicking the submit button.
        this.submitTarget.click();
        submitted = true;
      }
    } catch (e: unknown) {
      document.dispatchEvent(
        new CustomEvent("passkey:error", {
          detail: e,
        })
      );
    } finally {
      this.requesting = false;
      // After a passkey is chosen, the form is submitted and the page changes,
      // so the button stays disabled.
      if (this.hasButtonTarget && !submitted) {
        this.buttonTarget.disabled = false;
      }
      document.dispatchEvent(new CustomEvent(PASSKEY_MODAL_END));
    }
  }
}

export class AuthflowPasskeyCreationController extends Controller {
  static targets = ["button", "submit", "input"];
  static values = {
    options: String,
  };

  declare buttonTarget: HTMLButtonElement;
  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;
  declare hasButtonTarget: boolean;

  declare optionsValue: string;

  connect() {
    if (this.hasButtonTarget) {
      this.buttonTarget.disabled = false;
      // Disable the button if PublicKeyCredential is unavailable.
      if (!passkeyIsAvailable()) {
        this.buttonTarget.disabled = true;
        return;
      }
    }
  }

  create(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    this._create();
  }

  async _create() {
    if (!passkeyIsAvailable()) {
      document.dispatchEvent(
        new CustomEvent("passkey:error", {
          detail: new Error("passkey is not available"),
        })
      );
      return;
    }
    try {
      const optionsJSON = JSON.parse(this.optionsValue);
      const options = deserializeCreationOptions(optionsJSON);
      const rawResponse = await window.navigator.credentials.create(options);
      if (isPublicKeyCredentialWithAttestationResponse(rawResponse)) {
        const response = serializeAttestationResponse(rawResponse);
        const responseString = JSON.stringify(response);
        this.inputTarget.value = responseString;
        // It seems that we should use form.submit() to submit the form.
        // but form.submit() does NOT trigger submit event,
        // which is essential for XHR form submission to work.
        // Therefore, we emulate form submission here by clicking the submit button.
        this.submitTarget.click();
      }
    } catch (e: unknown) {
      document.dispatchEvent(
        new CustomEvent("passkey:error", {
          detail: e,
        })
      );
    }
  }
}

// Passkey autofill (WebAuthn conditional mediation) on the login ID input,
// which has "webauthn" in its autocomplete attribute.
//
// It was once disabled because it was buggy on iOS 16 Beta 4.
// The call navigator.credentials.get had a high possibility of resulting in
// DOMException with name NotAllowedError instantly.
// The exception caused subsequent calls to navigator.credentials.get to fail as well.
// So when the request keeps failing instantly, we stop retrying on this page.
//
// The options are the passkey request options of the authflow on the page.
// Every attempt reuses them, since the challenge is consumed only when the flow finishes.
//
// The lifecycle of the autofill
//
// start() begins a run, which makes a request, and again after each retry.
// In connect(), we call start(). In disconnect(), we call stop(), which ends the run
// and aborts the request that is pending.
// Since Stimulus will call connect(), we do not need to call start() if the request is aborted.
//
// Each request has its own AbortController, and it is aborted only while the request is pending.
// WebKit asks the platform to cancel once for every request made with a signal when the signal
// aborts, including requests that have ended, and a request made during a cancel waits for it.
//
// While AuthflowPasskeyRequestController shows the modal dialog,
// it dispatches PASSKEY_MODAL_START and PASSKEY_MODAL_END, and we stop() in between.
//
// When the request ends without a passkey, for example the user cancels Face ID,
// we start again after AUTOFILL_RETRY_DELAY_MS, so that the suggestions come back.
// The options are on the page, so a retry makes no network request and needs no backoff.
const AUTOFILL_RETRY_DELAY_MS = 1000;
// A request that ends within this time failed instantly, without the user doing anything.
const AUTOFILL_INSTANT_FAILURE_MS = 500;
// The number of instant failures in a row after which we stop retrying.
const AUTOFILL_MAX_INSTANT_FAILURES = 3;

// Errors that a retry cannot fix, such as a page that cannot claim the RP ID.
function isAutofillErrorFinal(err: unknown): boolean {
  return (
    err instanceof DOMException &&
    (err.name === "SecurityError" || err.name === "NotSupportedError")
  );
}

export class PasskeyAutofillController extends Controller {
  static targets = ["submit", "input"];
  static values = {
    options: String,
  };

  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;

  declare optionsValue: string;

  // The current run. stop() sets it to null, so that a run that is stopped while awaiting
  // neither starts a request nor schedules a retry.
  run: object | null = null;
  // Set only while a request is pending.
  requestAbortController: AbortController | null = null;
  retryTimer: ReturnType<typeof setTimeout> | null = null;
  instantFailures = 0;

  connect() {
    document.addEventListener(PASSKEY_MODAL_START, this.onModalStart);
    document.addEventListener(PASSKEY_MODAL_END, this.onModalEnd);
    window.addEventListener("pageshow", this.onPageShow);
    this.start();
  }

  disconnect() {
    document.removeEventListener(PASSKEY_MODAL_START, this.onModalStart);
    document.removeEventListener(PASSKEY_MODAL_END, this.onModalEnd);
    window.removeEventListener("pageshow", this.onPageShow);
    this.stop();
  }

  // A page restored from the back-forward cache is not connected again,
  // and its pending request is gone, so start again.
  onPageShow = (e: PageTransitionEvent) => {
    if (!e.persisted) {
      return;
    }
    this.stop();
    this.start();
  };

  onModalStart = () => {
    this.stop();
  };

  onModalEnd = () => {
    this.stop();
    this.start();
  };

  start() {
    if (!passkeyIsAvailable()) {
      return;
    }

    this.run = {};
    this.instantFailures = 0;
    void this.setupAutofill();
  }

  stop() {
    this.run = null;
    this.requestAbortController?.abort();
    this.requestAbortController = null;
    if (this.retryTimer != null) {
      clearTimeout(this.retryTimer);
      this.retryTimer = null;
    }
  }

  scheduleRetry() {
    this.retryTimer = setTimeout(() => {
      this.retryTimer = null;
      void this.setupAutofill();
    }, AUTOFILL_RETRY_DELAY_MS);
  }

  async setupAutofill() {
    const run = this.run;
    if (run == null) {
      return;
    }

    if (
      typeof PublicKeyCredential.isConditionalMediationAvailable !== "function"
    ) {
      return;
    }
    const available =
      await PublicKeyCredential.isConditionalMediationAvailable();
    if (!available || this.run !== run) {
      return;
    }

    const options = deserializeRequestOptions(JSON.parse(this.optionsValue));
    const requestAbortController = new AbortController();
    this.requestAbortController = requestAbortController;
    const t0 = new Date().getTime();
    try {
      const rawResponse = await window.navigator.credentials.get({
        ...options,
        mediation: "conditional",
        signal: requestAbortController.signal,
      });
      if (isPublicKeyCredentialWithAssertionResponse(rawResponse)) {
        const response = serializeAssertionResponse(rawResponse);
        const responseString = JSON.stringify(response);
        this.inputTarget.value = responseString;
        // It seems that we should use form.submit() to submit the form.
        // but form.submit() does NOT trigger submit event,
        // which is essential for XHR form submission to work.
        // Therefore, we emulate form submission here by clicking the submit button.
        // The page changes, so there is no retry.
        this.submitTarget.click();
        return;
      }
      if (this.run === run) {
        this.scheduleRetry();
      }
    } catch (e: unknown) {
      handleError(e);
      if (this.run !== run) {
        // Aborted. Let start() be called again.
        return;
      }
      if (isAutofillErrorFinal(e)) {
        return;
      }
      const t1 = new Date().getTime();
      if (t1 - t0 < AUTOFILL_INSTANT_FAILURE_MS) {
        this.instantFailures += 1;
        if (this.instantFailures >= AUTOFILL_MAX_INSTANT_FAILURES) {
          console.warn(
            "passkey autofill keeps failing instantly, stop retrying",
            e
          );
          return;
        }
      } else {
        this.instantFailures = 0;
      }
      this.scheduleRetry();
    } finally {
      if (this.requestAbortController === requestAbortController) {
        this.requestAbortController = null;
      }
    }
  }
}
