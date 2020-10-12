import { APIValidationError } from "./validation";
import { APIInvariantViolationError } from "./invariant";
import { APIDuplicatedIdentityError } from "./duplicated";
import { APIPasswordPolicyViolatedError } from "./password";

export type APIError =
  | APIValidationError
  | APIInvariantViolationError
  | APIDuplicatedIdentityError
  | APIPasswordPolicyViolatedError;

export function isAPIError(value?: { [key: string]: any }): value is APIError {
  if (value == null) {
    return false;
  }
  return "errorName" in value && "reason" in value;
}
