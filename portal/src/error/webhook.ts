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

export interface HookDeliveryTimeoutError {
  errorName: "InternalError";
  reason: "HookDeliveryTimeout";
}

export interface HookInvalidResponseError {
  errorName: "InternalError";
  reason: "HookInvalidResponse";
}
