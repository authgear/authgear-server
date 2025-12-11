export interface APIInvalidAccountStatusTransitionError {
  errorName: string;
  info: {
    cause: {
      kind: InvalidAccountStatusTransitionErrorInfoCauseKind;
    };
  };
  reason: "InvalidAccountStatusTransition";
}

export type InvalidAccountStatusTransitionErrorInfoCauseKind =
  | "AccountValidFromShouldBeBeforeAccountValidUntil"
  | "TemporarilyDisabledPeriodMissingUntilTimestamp"
  | "TemporarilyDisabledPeriodMissingFromTimestamp"
  | "TemporarilyDisabledFromShouldBeBeforeTemporarilyDisabledUntil"
  | "AccountValidFromShouldBeBeforeTemporarilyDisabledFrom"
  | "TemporarilyDisabledUntilShouldBeBeforeAccountValidUntil";
