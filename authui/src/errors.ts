// The original work is https://github.com/sindresorhus/is-network-error/blob/main/index.js
// The modifications are
// 1. Remove the error message checking for Node.js, since this code does not run on Node.js.
// 2. The original work also check error.stack === undefined on Safari >= 17, but in our observation,
//    error.stack IS NOT undefined, so we removed that checking.
// 3. The original work call Object.prototype.toString on error, and check if the string is "[object Error]",
//    we just use instanceof to check if it is an error.
const errorMessages = new Set([
  "network error", // Chrome
  "Failed to fetch", // Chrome
  "NetworkError when attempting to fetch resource.", // Firefox
  "The Internet connection appears to be offline.", // Safari 16
  "Load failed", // Safari 17+
  "Network request failed", // `cross-fetch`
]);

export function isNetworkError(error: unknown): boolean {
  const isValid =
    error &&
    error instanceof Error &&
    error.name === "TypeError" &&
    typeof error.message === "string";

  if (!isValid) {
    return false;
  }

  return errorMessages.has(error.message);
}
