/**
 * This code duplicates with authui/src/alert-message.ts. However, alert-message.ts cannot import other modules because it is commonjs script instead
 * Please help keep code in `setErrorMessage.ts` and `alert-message.ts` sync if you are to make any changes
 */

/**
 * Set error message for client-side display
 *
 * @param {string} id - ID of the error message, as defined in "__alert_message.html"
 */
export function setErrorMessage(id: string) {
  const e = new CustomEvent("alert-message:show-message", {
    detail: {
      id,
    },
  });
  document.dispatchEvent(e);
}
