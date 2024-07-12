import { Controller } from "@hotwired/stimulus";
import axios from "axios";
import { handleAxiosError, showErrorMessage } from "./messageBar";
import { base64DecToArr, base64EncArr } from "./base64";
import { base64URLToBase64, trimNewline, base64ToBase64URL } from "./base64url";
import { RetryEventTarget } from "./retry";

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

function serializeAttestationResponse(credential: PublicKeyCredential) {
  const response = credential.response as AuthenticatorAttestationResponse;

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

function serializeAssertionResponse(credential: PublicKeyCredential) {
  const response = credential.response as AuthenticatorAssertionResponse;
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

function isEmptyAllowCredentialsError(err: unknown) {
  return err instanceof DOMException && /allowCredentials/i.test(err.message);
}

function isSafariCancel(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "NotAllowedError" &&
    // "This request has been cancelled by the user."
    /cancel/i.test(err.message)
  );
}

function isSafariTimeout(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "HierarchyRequestError" &&
    // "The operation would yield an incorrect node tree."
    /node tree/i.test(err.message)
  );
}

function isOperationFailError(err: unknown) {
  // This happens when the user chooses to use Android phone to scan QR code.
  // And the passkey is not found on that Android phone.
  // If security key is used, the error message is shown in the modal dialog and the dialog will not close.
  // Thus we cannot detect security key error.

  // This error also happens when using passkey api in webview
  return (
    err instanceof DOMException &&
    err.name === "NotAllowedError" &&
    // "Operation failed."
    /operation.*fail/i.test(err.message)
  );
}

function isChromeCancelOrTimeout(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "NotAllowedError" &&
    // "The operation either timed out or was not allowed. See: https://www.w3.org/TR/webauthn-2/#sctn-privacy-considerations-client."
    /time.*out/i.test(err.message)
  );
}

function isFirefoxCancel(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "AbortError" &&
    // "The operation was aborted. "
    /operation.*abort/i.test(err.message)
  );
}

function isAbortControllerError(err: unknown) {
  return err instanceof DOMException && err.name === "AbortError";
}

function isFirefoxSecurityKeyError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // "An attempt was made to use an object that is not, or is no longer, usable"
    /no.*usable/i.test(err.message)
  );
}

function isSafariDuplicateError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // "At least one credential matches an entry of the excludeCredentials list in the platform attached authenticator."
    /excludeCredentials/i.test(err.message)
  );
}

function isChromeDuplicateError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // "The user attempted to register an authenticator that contains one of the credentials already registered with the relying party."
    /already register/i.test(err.message)
  );
}

function handleError(err: unknown) {
  console.error(err);

  const errorThatCanSimplyBeIgnored = [
    isSafariCancel,
    isSafariTimeout,
    isChromeCancelOrTimeout,
    isFirefoxCancel,
    // Firefox timeout was not observed.
    isAbortControllerError,
  ];
  for (const p of errorThatCanSimplyBeIgnored) {
    if (p(err)) {
      return;
    }
  }

  if (isOperationFailError(err)) {
    showErrorMessage("error-message-no-passkey");
    return;
  }

  if (isFirefoxSecurityKeyError(err)) {
    showErrorMessage("error-message-no-passkey");
    return;
  }

  if (isEmptyAllowCredentialsError(err)) {
    showErrorMessage("error-message-passkey-empty-allow-credentials");
  }

  if (isSafariDuplicateError(err) || isChromeDuplicateError(err)) {
    showErrorMessage("error-message-passkey-duplicate");
  }

  return false;
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
        if (rawResponse instanceof PublicKeyCredential) {
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
        if (rawResponse instanceof PublicKeyCredential) {
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
    const errorThatCanSimplyBeIgnored = [
      isSafariCancel,
      isSafariTimeout,
      isChromeCancelOrTimeout,
      isFirefoxCancel,
      // Firefox timeout was not observed.
      isAbortControllerError,
    ];
    for (const p of errorThatCanSimplyBeIgnored) {
      if (p(err)) {
        return;
      }
    }
    let errMessage = "data-passkey-not-supported";
    if (isOperationFailError(err)) {
      errMessage = "data-invalid-passkey-or-not-supported";
    }

    if (isFirefoxSecurityKeyError(err)) {
      errMessage = "data-no-passkey";
    }

    if (isSafariDuplicateError(err) || isChromeDuplicateError(err)) {
      errMessage = "data-passkey-duplicate";
    }
    document.dispatchEvent(
      new CustomEvent("alert-message:show-message", {
        detail: { id: errMessage },
      })
    );
  };
}

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

  connect() {
    this.buttonTarget.disabled = false;
    // Disable the button if PublicKeyCredential is unavailable.
    if (!passkeyIsAvailable()) {
      if (this.hasButtonTarget) {
        this.buttonTarget.disabled = true;
      }
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
      const options = deserializeRequestOptions(optionsJSON);
      const rawResponse = await window.navigator.credentials.get(options);
      if (rawResponse instanceof PublicKeyCredential) {
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
      document.dispatchEvent(
        new CustomEvent("passkey:error", {
          detail: e,
        })
      );
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
    this.buttonTarget.disabled = false;
    // Disable the button if PublicKeyCredential is unavailable.
    if (!passkeyIsAvailable()) {
      if (this.hasButtonTarget) {
        this.buttonTarget.disabled = true;
      }
      return;
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
      if (rawResponse instanceof PublicKeyCredential) {
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

// TODO(passkey): autofill is buggy on iOS 16 Beta 4.
// The call navigator.credentials.get will have a high possibility of resulting in
// DOMException with name NotAllowedError instantly.
// The exception WILL cause subsequent call to navigator.credentials.get to fail as well.
// So autofill is disabled for now.
//
// The lifecycle of the autofill
//
// setupAutofill() creates a pending promise to receive the result of the autofill.
// The promise can be aborted with AbortController.
// In disconnect(), we abort the promise.
// In connect(), we call setupAutofill().
// Since Stimulus will call connect(), we do not need to call setupAutofill if the promise is aborted.
export class PasskeyAutofillController extends Controller {
  static targets = ["submit", "input"];

  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;

  abortController: AbortController | null = null;
  retryEventTarget: RetryEventTarget | null = null;

  connect() {
    if (!passkeyIsAvailable()) {
      return;
    }

    this.abortController = new AbortController();
    this.retryEventTarget = new RetryEventTarget({
      abortController: this.abortController,
    });
    this.retryEventTarget.addEventListener("retry", () => {
      this.setupAutofill();
    });

    this.setupAutofill();
  }

  disconnect() {
    this.abortController?.abort();
    this.abortController = null;
    this.retryEventTarget = null;
  }

  async setupAutofill() {
    if (
      typeof PublicKeyCredential.isConditionalMediationAvailable === "function"
    ) {
      const available =
        await PublicKeyCredential.isConditionalMediationAvailable();

      if (available) {
        try {
          const params = new URLSearchParams();
          params.set("conditional", "true");
          const resp = await axios("/_internals/passkey/request_options", {
            method: "post",
            data: params,
          });
          this.retryEventTarget?.markSuccess();
          const options = deserializeRequestOptions(resp.data.result);
          const t0 = new Date().getTime();
          try {
            const signal = this.abortController?.signal;
            const rawResponse = await window.navigator.credentials.get({
              ...options,
              signal,
            });
            if (rawResponse instanceof PublicKeyCredential) {
              const response = serializeAssertionResponse(rawResponse);
              const responseString = JSON.stringify(response);
              this.inputTarget.value = responseString;
              // It seems that we should use form.submit() to submit the form.
              // but form.submit() does NOT trigger submit event,
              // which is essential for XHR form submission to work.
              // Therefore, we emulate form submission here by clicking the submit button.
              this.submitTarget.click();
            }
            this.retryEventTarget?.scheduleRetry();
          } catch (e: unknown) {
            const t1 = new Date().getTime();
            if (
              e instanceof DOMException &&
              e.name === "NotAllowedError" &&
              t1 - t0 < 500
            ) {
              console.warn("passkey autofill weird error detected", e);
            }
            if (e instanceof DOMException && e.name === "AbortError") {
              // Aborted. Let connect() to be called again.
            } else {
              this.retryEventTarget?.scheduleRetry();
            }
            handleError(e);
          }
        } catch (e: unknown) {
          handleAxiosError(e);
          this.retryEventTarget?.scheduleRetry();
        }
      }
    }
  }
}
