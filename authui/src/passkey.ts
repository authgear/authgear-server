import { Controller } from "@hotwired/stimulus";
import axios from "axios";
import { handleAxiosError } from "./messageBar";
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

function handleError(err: unknown) {
  if (err instanceof DOMException && err.name === "NotAllowedError") {
    return;
  }
  // Abort
  if (err instanceof DOMException && err.name === "AbortError") {
    return;
  }

  console.error(err);
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
      const resp = await axios("/passkey/creation_options", {
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

  use(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    this._use();
  }

  async _use() {
    try {
      const resp = await axios("/passkey/request_options", {
        method: "post",
        headers: {
          "content-type": "application/json; charset=utf-8",
        },
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
      // @ts-expect-error
      typeof PublicKeyCredential.isConditionalMediationAvailable === "function"
    ) {
      const available =
        // @ts-expect-error
        await PublicKeyCredential.isConditionalMediationAvailable();

      if (available) {
        try {
          const params = new URLSearchParams();
          params.set("conditional", "true");
          const resp = await axios("/passkey/request_options", {
            method: "post",
            data: params,
          });
          this.retryEventTarget?.markSuccess();
          const options = deserializeRequestOptions(resp.data.result);
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
            handleError(e);
            if (e instanceof DOMException && e.name === "AbortError") {
              // Aborted. Let connect() to be called again.
            } else {
              this.retryEventTarget?.scheduleRetry();
            }
          }
        } catch (e: unknown) {
          handleAxiosError(e);
          this.retryEventTarget?.scheduleRetry();
        }
      }
    }
  }
}
