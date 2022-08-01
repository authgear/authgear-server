import { Controller } from "@hotwired/stimulus";
import axios from "axios";
import { handleAxiosError } from "./messageBar";
import { base64DecToArr, base64EncArr } from "./base64";
import { base64URLToBase64, trimNewline, base64ToBase64URL } from "./base64url";

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
  // Cancel
  if (err instanceof DOMException && err.name === "NotAllowedError") {
    return;
  }

  console.error(err);
}

export class PasskeyCreationController extends Controller {
  static targets = ["button", "submit", "input"];
  static values = {
    preferred: Boolean,
  };

  declare buttonTarget: HTMLButtonElement;
  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;
  declare preferredValue: boolean;

  connect() {
    // Disable the button if PublicKeyCredential is unavailable.
    if (!passkeyIsAvailable()) {
      this.buttonTarget.disabled = true;
      return;
    }

    if (this.preferredValue) {
      this._create();
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

export class PasskeyRequestController extends Controller {
  static targets = ["button", "submit", "input"];
  static values = {
    preferred: Boolean,
  };

  declare buttonTarget: HTMLButtonElement;
  declare submitTarget: HTMLButtonElement;
  declare inputTarget: HTMLInputElement;
  declare preferredValue: boolean;

  connect() {
    // Disable the button if PublicKeyCredential is unavailable.
    if (!passkeyIsAvailable()) {
      this.buttonTarget.disabled = true;
      return;
    }

    if (this.preferredValue) {
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
