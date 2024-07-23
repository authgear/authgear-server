/**
 * Dispatch a custom event to set captcha verified with success token
 *
 * @param {string} token - success token
 */
export function dispatchBotProtectionEventVerified(token: string) {
  document.dispatchEvent(
    new CustomEvent("bot-protection:verify-success", {
      detail: {
        token,
      },
    })
  );
}

/**
 * Dispatch a custom event to set captcha failed
 *
 * @param {string | undefined} errMsg - error message
 */
export function dispatchBotProtectionEventFailed(errMsg?: string) {
  document.dispatchEvent(
    new CustomEvent("bot-protection:verify-failed", {
      detail: {
        errMsg,
      },
    })
  );
}

/**
 * Dispatch a custom event to set captcha expired
 *
 * @param {string | undefined} token - expired token
 */
export function dispatchBotProtectionEventExpired(token?: string) {
  document.dispatchEvent(
    new CustomEvent("bot-protection:verify-expired", {
      detail: {
        token,
      },
    })
  );
}
