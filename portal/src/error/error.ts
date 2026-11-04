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

// RateLimitedError is what the server returns when a rate limit is exceeded.
// It arrives with HTTP 429 and, for a request rejected before it reaches the
// GraphQL executor, outside the GraphQL errors array -- see parseRawError.
export interface RateLimitedError {
  errorName: "TooManyRequest";
  reason: "RateLimited";
}

export interface ServiceUnavailableError {
  errorName: "ServiceUnavailable";
  reason: "ServiceUnavailable";
}

// apierrors.NewInvalid uses the kind as the reason.
export interface GenericInvalidError {
  errorName: "Invalid";
  reason: "Invalid";
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

export interface APIDenoRunError {
  errorName: string;
  reason: "DenoRunError";
}

export interface APIHookDeliveryUnknownFailureError {
  errorName: string;
  reason: "HookDeliveryUnknownFailure";
}

export interface APIUsageLimitExceededError {
  errorName: string;
  reason: "UsageLimitExceeded";
  info: {
    usage_name: string;
    period?: string;
  };
}

export interface APIMFAGracePeriodInvalidError {
  errorName: string;
  reason: "MFAGracePeriodInvalid";
}

export interface APIScopeDuplicateError {
  errorName: string;
  reason: "ScopeDuplicate";
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
    | "SMSGatewayRateLimited"
    | "SMSGatewayTimeout"
    | "SMSGatewayUnsupportedRequest";
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
  | GenericInvalidError
  | LocalError
  | TooManyRequestError
  | RateLimitedError
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
  | APIDenoRunError
  | APIHookDeliveryUnknownFailureError
  | APIUsageLimitExceededError
  | APIMFAGracePeriodInvalidError
  | APIScopeDuplicateError
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
