export interface HookDisallowedError {
  errorName: "Forbidden";
  reason: "HookDisallowed";
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
