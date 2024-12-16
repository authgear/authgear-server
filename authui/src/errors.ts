// Modified from https://github.com/sindresorhus/is-network-error/blob/main/index.js
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
