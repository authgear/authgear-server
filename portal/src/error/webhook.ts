export interface WebHookDisallowedError {
  errorName: "Forbidden";
  reason: "WebHookDisallowed";
}

export interface WebHookDeliveryTimeoutError {
  errorName: "InternalError";
  reason: "WebHookDeliveryTimeout";
}

export interface WebHookInvalidResponseError {
  errorName: "InternalError";
  reason: "WebHookInvalidResponse";
}
