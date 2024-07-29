// These are meaningful errors types for handling, which originated from actual cloudflare turnstile error codes
export enum CloudflareTurnstileError {
  InvalidSiteKey = "invalid-sitekey",
  DomainNotAllowed = "domain-not-allowed",
  UnusualVisitorBehavior = "unusual-visitor-behavior",
  UnsupportedBrowser = "unsupported-browser",
  Timeout = "timeout",
  Unknown = "unknown",
}

export interface CloudflareTurnstileErrorCodeParseResult {
  error: CloudflareTurnstileError;
  shouldRetry: boolean;
  shouldDisplayErrMsg: boolean;
}

const CLOUDFLARE_TURNSTILE_ERROR_INVALID_SITE_KEY: CloudflareTurnstileErrorCodeParseResult =
  {
    error: CloudflareTurnstileError.InvalidSiteKey,
    shouldRetry: false,
    shouldDisplayErrMsg: false, // cloudflare will render the error message inside its widget
  };

const CLOUDFLARE_TURNSTILE_ERROR_DOMAIN_NOT_ALLOWED: CloudflareTurnstileErrorCodeParseResult =
  {
    error: CloudflareTurnstileError.DomainNotAllowed,
    shouldRetry: false,
    shouldDisplayErrMsg: false, // cloudflare will render the error message inside its widget
  };
const CLOUDFLARE_TURNSTILE_ERROR_UNUSUAL_VISITOR_BEHAVIOR: CloudflareTurnstileErrorCodeParseResult =
  {
    error: CloudflareTurnstileError.UnusualVisitorBehavior,
    shouldRetry: true,
    shouldDisplayErrMsg: false,
  };
const CLOUDFLARE_TURNSTILE_ERROR_UNSUPPORTED_BROWSER: CloudflareTurnstileErrorCodeParseResult =
  {
    error: CloudflareTurnstileError.UnsupportedBrowser,
    shouldRetry: false,
    shouldDisplayErrMsg: false, // cloudflare will render the error message inside its widget
  };
const CLOUDFLARE_TURNSTILE_ERROR_TIMEOUT: CloudflareTurnstileErrorCodeParseResult =
  {
    error: CloudflareTurnstileError.Timeout,
    shouldRetry: true,
    shouldDisplayErrMsg: false, // auto retry, no need err message
  };
const CLOUDFLARE_TURNSTILE_ERROR_UNKNOWN: CloudflareTurnstileErrorCodeParseResult =
  {
    error: CloudflareTurnstileError.Unknown,
    shouldRetry: false,
    shouldDisplayErrMsg: true,
  };

/**
 * @see https://developers.cloudflare.com/turnstile/troubleshooting/client-side-errors/error-codes/
 */
export function parseCloudflareTurnstileErrorCode(
  errCode: string
): CloudflareTurnstileErrorCodeParseResult {
  switch (errCode) {
    case "110100":
    case "110110": {
      return CLOUDFLARE_TURNSTILE_ERROR_INVALID_SITE_KEY;
    }

    case "110200": {
      return CLOUDFLARE_TURNSTILE_ERROR_DOMAIN_NOT_ALLOWED;
    }

    case "110500": {
      return CLOUDFLARE_TURNSTILE_ERROR_UNSUPPORTED_BROWSER;
    }

    default: {
      if (errCode.startsWith("1106")) {
        return CLOUDFLARE_TURNSTILE_ERROR_TIMEOUT;
      }
      if (errCode.startsWith("300") || errCode.startsWith("600")) {
        return CLOUDFLARE_TURNSTILE_ERROR_UNUSUAL_VISITOR_BEHAVIOR;
      }
      return CLOUDFLARE_TURNSTILE_ERROR_UNKNOWN;
    }
  }
}
