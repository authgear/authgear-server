export function isEmptyAllowCredentialsError(err: unknown) {
  return err instanceof DOMException && /allowCredentials/i.test(err.message);
}

export function isSafariCancel(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "NotAllowedError" &&
    // "This request has been cancelled by the user."
    /cancel/i.test(err.message)
  );
}

export function isSafariTimeout(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "HierarchyRequestError" &&
    // "The operation would yield an incorrect node tree."
    /node tree/i.test(err.message)
  );
}

export function isOperationFailError(err: unknown) {
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

export function isChromeCancelOrTimeout(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "NotAllowedError" &&
    // "The operation either timed out or was not allowed. See: https://www.w3.org/TR/webauthn-2/#sctn-privacy-considerations-client."
    /time.*out/i.test(err.message)
  );
}

export function isFirefoxCancel(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "AbortError" &&
    // "The operation was aborted. "
    /operation.*abort/i.test(err.message)
  );
}

export function isAbortControllerError(err: unknown) {
  return err instanceof DOMException && err.name === "AbortError";
}

export function isFirefoxSecurityKeyError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // "An attempt was made to use an object that is not, or is no longer, usable"
    /no.*usable/i.test(err.message)
  );
}

export function isSafariDuplicateError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // "At least one credential matches an entry of the excludeCredentials list in the platform attached authenticator."
    /excludeCredentials/i.test(err.message)
  );
}

export function isChromeDuplicateError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // "The user attempted to register an authenticator that contains one of the credentials already registered with the relying party."
    /already register/i.test(err.message)
  );
}
