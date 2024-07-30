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
import { APIBadNFTCollectionError, APIAlchemyProtocolError } from "./web3";
import type { ParsedAPIError } from "./parse";
import { APIResourceUpdateConflictError } from "./resourceUpdateConflict";

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
}

export interface LocalError {
  errorName: "__local";
  reason: "__local";
  info: {
    error: ParsedAPIError;
  };
}

export interface APIDenoCheckError {
  errorName: string;
  reason: "DenoCheckError";
}

export interface APIRoleDuplicateKeyError {
  errorName: string;
  reason: "RoleDuplicateKey";
}

export interface APIGroupDuplicateKeyError {
  errorName: string;
  reason: "GroupDuplicateKey";
}

export interface APIEmailIdentityNotFoundError {
  errorName: string;
  reason: "EmailIdentityNotFound";
}

export type APIError = { message?: string } & (
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
  | APIBadNFTCollectionError
  | APIAlchemyProtocolError
  | APIDenoCheckError
  | APIResourceUpdateConflictError
  | APIRoleDuplicateKeyError
  | APIGroupDuplicateKeyError
  | APIEmailIdentityNotFoundError
);

export function isAPIError(value: unknown): value is APIError {
  return (
    typeof value === "object" &&
    !!value &&
    "errorName" in value &&
    "reason" in value
  );
}
