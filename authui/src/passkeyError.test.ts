import { describe, it, expect } from "@jest/globals";
import { getAuthflowV2PasskeyErrorMessageID } from "./passkeyError";

describe("getAuthflowV2PasskeyErrorMessageID", () => {
  const cases: [string, unknown, string | null][] = [
    [
      "iOS 17 Safari cancel",
      new DOMException(
        "The operation couldn’t be completed. (com.apple.AuthenticationServices.AuthorizationError error 1001.)",
        "NotAllowedError"
      ),
      null,
    ],
    [
      "older Safari cancel",
      new DOMException(
        "This request has been cancelled by the user.",
        "NotAllowedError"
      ),
      null,
    ],
    [
      "Chrome cancel or timeout",
      new DOMException(
        "The operation either timed out or was not allowed. See: https://www.w3.org/TR/webauthn-2/#sctn-privacy-considerations-client.",
        "NotAllowedError"
      ),
      null,
    ],
    [
      "Firefox cancel",
      new DOMException("The operation was aborted. ", "AbortError"),
      null,
    ],
    [
      "Safari timeout",
      new DOMException(
        "The operation would yield an incorrect node tree.",
        "HierarchyRequestError"
      ),
      null,
    ],
    [
      "operation failed",
      new DOMException("Operation failed.", "NotAllowedError"),
      "data-invalid-passkey-or-not-supported",
    ],
    [
      "Firefox security key",
      new DOMException(
        "An attempt was made to use an object that is not, or is no longer, usable",
        "InvalidStateError"
      ),
      "data-no-passkey",
    ],
    [
      "Safari duplicate",
      new DOMException(
        "At least one credential matches an entry of the excludeCredentials list in the platform attached authenticator.",
        "InvalidStateError"
      ),
      "data-passkey-duplicate",
    ],
    [
      "Chrome duplicate",
      new DOMException(
        "The user attempted to register an authenticator that contains one of the credentials already registered with the relying party.",
        "InvalidStateError"
      ),
      "data-passkey-duplicate",
    ],
    [
      "browser rejects empty allowCredentials",
      new DOMException("allowCredentials must not be empty", "NotAllowedError"),
      "data-passkey-not-supported",
    ],
    [
      "passkey API unavailable",
      new Error("passkey is not available"),
      "data-passkey-not-supported",
    ],
    [
      "NotSupportedError",
      new DOMException("Not supported", "NotSupportedError"),
      "data-passkey-not-supported",
    ],
  ];

  it.each(cases)("%s", (_name, err, expected) => {
    expect(getAuthflowV2PasskeyErrorMessageID(err)).toBe(expected);
  });
});
