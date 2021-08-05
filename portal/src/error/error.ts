import { APIValidationError } from "./validation";
import { APIInvariantViolationError } from "./invariant";
import { APIPasswordPolicyViolatedError } from "./password";
import { APIForbiddenError } from "./forbidden";
import {
  APIDuplicatedDomainError,
  APIDomainVerifiedError,
  APIDomainNotFoundError,
  APIDomainNotCustomError,
  APIDomainVerificationFailedError,
  APIInvalidDomainError,
} from "./domain";
import {
  APIDuplicatedCollaboratorInvitationError,
  APICollaboratorSelfDeletionError,
  APICollaboratorDuplicateError,
  APICollaboratorInvitationInvalidCodeError,
  APICollaboratorInvitationInvalidEmailError,
} from "./collaborator";
import {
  APIDuplicatedAppIDError,
  APIInvalidAppIDError,
  APIReservedAppIDError,
} from "./apps";
import {
  APIResourceNotFoundError,
  APIResourceTooLargeError,
} from "./resources";

export interface NetworkError {
  errorName: "NetworkFailed";
  reason: "NetworkFailed";
}

export interface UnknownError {
  errorName: "Unknown";
  reason: "Unknown";
  info: {
    message: string;
  };
}

export type APIError =
  | NetworkError
  | UnknownError
  | APIValidationError
  | APIInvariantViolationError
  | APIPasswordPolicyViolatedError
  | APIForbiddenError
  | APIDuplicatedDomainError
  | APIDomainVerifiedError
  | APIDomainNotFoundError
  | APIDomainNotCustomError
  | APIDomainVerificationFailedError
  | APIInvalidDomainError
  | APIDuplicatedCollaboratorInvitationError
  | APICollaboratorSelfDeletionError
  | APICollaboratorInvitationInvalidCodeError
  | APICollaboratorInvitationInvalidEmailError
  | APICollaboratorDuplicateError
  | APIDuplicatedAppIDError
  | APIInvalidAppIDError
  | APIReservedAppIDError
  | APIResourceNotFoundError
  | APIResourceTooLargeError;

export function isAPIError(value: unknown): value is APIError {
  return (
    typeof value === "object" &&
    !!value &&
    "errorName" in value &&
    "reason" in value
  );
}
