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
  APIUnsupportedImageFileError,
} from "./resources";
import {
  WebHookDisallowedError,
  WebHookDeliveryTimeoutError,
  WebHookInvalidResponseError,
} from "./webhook";
import { APIBadNFTCollectionError } from "./web3";
import type { ParsedAPIError } from "./parse";

export interface NetworkError {
  errorName: "NetworkFailed";
  reason: "NetworkFailed";
}

export interface RequestEntityTooLargeError {
  errorName: "RequestEntityTooLarge";
  reason: "RequestEntityTooLarge";
}

export interface TooManyRequestError {
  errorName: "TooManyRequest";
  reason: "TooManyRequest";
}

export interface ServiceUnavailableError {
  errorName: "ServiceUnavailable";
  reason: "ServiceUnavailable";
}

export interface UnknownError {
  errorName: "Unknown";
  reason: "Unknown";
  info: {
    message: string;
  };
}

export interface LocalError {
  errorName: "__local";
  reason: "__local";
  info: {
    error: ParsedAPIError;
  };
}

export type APIError =
  | NetworkError
  | RequestEntityTooLargeError
  | UnknownError
  | LocalError
  | TooManyRequestError
  | ServiceUnavailableError
  | WebHookDisallowedError
  | WebHookDeliveryTimeoutError
  | WebHookInvalidResponseError
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
  | APIResourceTooLargeError
  | APIUnsupportedImageFileError
  | APIBadNFTCollectionError;

export function isAPIError(value: unknown): value is APIError {
  return (
    typeof value === "object" &&
    !!value &&
    "errorName" in value &&
    "reason" in value
  );
}
