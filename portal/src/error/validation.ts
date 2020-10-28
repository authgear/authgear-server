export interface APIValidationError {
  errorName: string;
  reason: "ValidationFailed";
  info: ValidationFailedErrorInfo;
}

export interface ValidationFailedErrorInfo {
  causes: ValidationFailedErrorInfoCause[];
}

export type ValidationFailedErrorInfoCause =
  | RequiredErrorCause
  | GeneralErrorCause
  | FormatErrorCause
  | MinItemsErrorCause
  | MinimumErrorCause
  | MaximumErrorCause
  | MinLengthErrorCause
  | MaxLengthErrorCause;

export interface RequiredErrorCause {
  location: string;
  kind: "required";
  details: {
    actual: string[];
    expected: string[];
    missing: string[];
  };
}

export interface GeneralErrorCause {
  location: string;
  kind: "general";
  details: {
    msg: string;
  };
}

export interface FormatErrorCause {
  location: string;
  kind: "format";
  details: {
    format: string;
  };
}

export interface MinItemsErrorCause {
  location: string;
  kind: "minItems";
  details: {
    actual: number;
    expected: number;
  };
}

export interface MinimumErrorCause {
  location: string;
  kind: "minimum";
  details: {
    actual: number;
    minimum: number;
  };
}

export interface MaximumErrorCause {
  location: string;
  kind: "maximum";
  details: {
    actual: number;
    maximum: number;
  };
}

export interface MinLengthErrorCause {
  location: string;
  kind: "minLength";
  details: {
    actual: number;
    expected: number;
  };
}

export interface MaxLengthErrorCause {
  location: string;
  kind: "maxLength";
  details: {
    actual: number;
    expected: number;
  };
}
