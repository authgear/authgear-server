import { APIValidationError } from "./validation";
import { APIInvariantViolationError } from "./invariant";
import { APIInvalidAccountStatusTransitionError } from "./accountStatus";
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
  HookDisallowedError,
  HookDeliveryTimeoutError,
  HookInvalidResponseError,
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

export interface ForbiddenError {
  errorName: "Forbidden";
  reason: "Forbidden";
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

export interface APIResourceDuplicateURIError {
  errorName: string;
  reason: "ResourceDuplicateURI";
}

export interface APISendPasswordNoTargetError {
  errorName: string;
  reason: "SendPasswordNoTarget";
}

export interface APIAuthenticatorNotFoundError {
  errorName: string;
  reason: "AuthenticatorNotFound";
}

export interface APISMSGatewayError {
  errorName: string;
  reason:
    | "SMSGatewayInvalidPhoneNumber"
    | "SMSGatewayAuthenticationFailed"
    | "SMSGatewayDeliveryRejected"
    | "SMSGatewayRateLimited";
  info: {
    ProviderErrorCode: string;
    ProviderName: string;
    Description: string;
  };
}

export interface APISMTPTestFailedError {
  errorName: string;
  reason: "SMTPTestFailed";
  message: string;
}

export type APIError = { message?: string } & (
  | NetworkError
  | RequestEntityTooLargeError
  | ForbiddenError
  | UnknownError
  | LocalError
  | TooManyRequestError
  | ServiceUnavailableError
  | HookDisallowedError
  | HookDeliveryTimeoutError
  | HookInvalidResponseError
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
  | APISendPasswordNoTargetError
  | APIAuthenticatorNotFoundError
  | APISMSGatewayError
  | APISMTPTestFailedError
  | APIResourceDuplicateURIError
  | APIInvalidAccountStatusTransitionError
);

export function isAPIError(value: unknown): value is APIError {
  return (
    typeof value === "object" &&
    !!value &&
    "errorName" in value &&
    "reason" in value
  );
}
