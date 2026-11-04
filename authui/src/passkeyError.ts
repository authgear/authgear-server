export function isEmptyAllowCredentialsError(err: unknown) {
  return err instanceof DOMException && /allowCredentials/i.test(err.message);
}

function isSafariTimeout(err: unknown) {
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

export function isFirefoxSecurityKeyError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // "An attempt was made to use an object that is not, or is no longer, usable"
    /no.*usable/i.test(err.message)
  );
}

export function isDuplicateError(err: unknown) {
  return (
    err instanceof DOMException &&
    err.name === "InvalidStateError" &&
    // Safari: "At least one credential matches an entry of the excludeCredentials list in the platform attached authenticator."
    // Chrome: "The user attempted to register an authenticator that contains one of the credentials already registered with the relying party."
    (/excludeCredentials/i.test(err.message) ||
      /already register/i.test(err.message))
  );
}

// The WebAuthn spec deliberately reports cancel, timeout and "not allowed" as
// the same NotAllowedError, and the message differs across browsers and
// versions (e.g. iOS 17 Safari says "...AuthorizationError error 1001."),
// so do not rely on the message to recognize a cancel.
export function isCancelOrTimeout(err: unknown) {
  if (!(err instanceof DOMException)) {
    return false;
  }
  switch (err.name) {
    case "NotAllowedError":
      return !isOperationFailError(err) && !isEmptyAllowCredentialsError(err);
    // Firefox cancel, or our own AbortController.
    case "AbortError":
      return true;
    default:
      return isSafariTimeout(err);
  }
}

// Returns the data attribute of the alert message to show, or null to show nothing.
export function getAuthflowV2PasskeyErrorMessageID(
  err: unknown
): string | null {
  if (isCancelOrTimeout(err)) {
    return null;
  }
  if (isOperationFailError(err)) {
    return "data-invalid-passkey-or-not-supported";
  }
  if (isFirefoxSecurityKeyError(err)) {
    return "data-no-passkey";
  }
  if (isDuplicateError(err)) {
    return "data-passkey-duplicate";
  }
  return "data-passkey-not-supported";
}
