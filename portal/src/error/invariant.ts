export interface APIInvariantViolationError {
  errorName: string;
  info: {
    cause: {
      kind: InvariantViolatedErrorInfoCauseKind;
    };
    // Set on DuplicatedIdentity.
    IdentityTypeIncoming?: string;
    LoginIDTypeIncoming?: string;
  };
  reason: "InvariantViolated";
}

export type InvariantViolatedErrorInfoCauseKind =
  | "RemoveLastIdentity"
  | "DuplicatedIdentity";
