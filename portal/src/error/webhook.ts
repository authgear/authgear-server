export interface WebHookDisallowedError {
  errorName: "Forbidden";
  reason: "WebHookDisallowed";
  info?: {
    reasons?: {
      title?: string;
      reason?: string;
    }[];
  };
}

export interface WebHookDeliveryTimeoutError {
  errorName: "InternalError";
  reason: "WebHookDeliveryTimeout";
}

export interface WebHookInvalidResponseError {
  errorName: "InternalError";
  reason: "WebHookInvalidResponse";
}
