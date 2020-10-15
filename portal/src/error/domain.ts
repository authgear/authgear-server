export interface APIDuplicatedDomainError {
  errorName: string;
  reason: "DuplicatedDomain";
}

export interface APIDomainVerifiedError {
  errorName: string;
  reason: "DomainVerified";
}

export interface APIDomainNotFoundError {
  errorName: string;
  reason: "DomainNotFound";
}

export interface APIDomainNotCustomError {
  errorName: string;
  reason: "DomainNotCustom";
}

export interface APIDomainVerificationFailedError {
  errorName: string;
  reason: "DomainVerificationFailed";
}

export interface APIInvalidDomainError {
  errorName: string;
  reason: "InvalidDomain";
}
