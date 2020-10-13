export interface APIPasswordPolicyViolatedError {
  errorName: string;
  reason: "PasswordPolicyViolated";
  info: {
    causes: PasswordPolicyViolatedErrorCause[];
  };
}

export interface PasswordPolicyViolatedErrorCause {
  Name: string;
  Info: unknown;
}
