export interface APIInvariantViolationError {
  errorName: string;
  info: {
    cause: {
      kind: InvariantViolatedErrorInfoCauseKind;
    };
  };
  reason: "InvariantViolated";
}

export type InvariantViolatedErrorInfoCauseKind = "DuplicatedAuthenticator";
