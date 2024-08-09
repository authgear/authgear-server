import { APIError } from "./error";
import { makeLocalValidationError } from "./validation";
import { PasswordPolicyConfig } from "../types";
import { GuessableLevel, zxcvbnGuessableLevel } from '../util/zxcvbn';

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

export function checkPasswordPolicy(
  passwordPolicy: PasswordPolicyConfig,
  password: string,
  level: GuessableLevel
): Partial<Record<keyof PasswordPolicyConfig, boolean>> {
  const isPolicySatisfied: Partial<
    Record<keyof PasswordPolicyConfig, boolean>
  > = {};
  if (password.length === 0) {
    return isPolicySatisfied;
  }

  if (passwordPolicy.min_length != null) {
    isPolicySatisfied.min_length = password.length >= passwordPolicy.min_length;
  }
  if (passwordPolicy.lowercase_required === true) {
    isPolicySatisfied.lowercase_required = /[a-z]/.test(password);
  }
  if (passwordPolicy.uppercase_required === true) {
    isPolicySatisfied.uppercase_required = /[A-Z]/.test(password);
  }
  if (passwordPolicy.alphabet_required === true) {
    isPolicySatisfied.alphabet_required = /[a-zA-Z]/.test(password);
  }
  if (passwordPolicy.digit_required === true) {
    isPolicySatisfied.digit_required = /\d/.test(password);
  }
  if (passwordPolicy.symbol_required === true) {
    // treat all character which is not alphanumeric as symbol
    isPolicySatisfied.symbol_required = /[^a-zA-Z0-9]/.test(password);
  }
  if (passwordPolicy.minimum_guessable_level != null) {
    isPolicySatisfied.minimum_guessable_level =
      level >= passwordPolicy.minimum_guessable_level;
  }

  return isPolicySatisfied;
}

export function isPasswordValid(
  passwordPolicy: PasswordPolicyConfig,
  password: string | undefined,
  level: GuessableLevel
): boolean {
  if (password == null) {
    return false;
  }
  const isPolicySatisfied = checkPasswordPolicy(
    passwordPolicy,
    password,
    level
  );
  return Object.values(isPolicySatisfied).every(Boolean);
}

export function validatePassword(
  password: string,
  policy: PasswordPolicyConfig,
  confirmPassword?: string
): APIError | null {
  if (confirmPassword != null && password !== confirmPassword) {
    return makeLocalValidationError([
      { location: "/confirm_password", messageID: "errors.password-mismatch" },
    ]);
  }

  const guessableLevel = zxcvbnGuessableLevel(password);
  const passwordValid = isPasswordValid(policy, password, guessableLevel);
  if (!passwordValid) {
    return makeLocalValidationError([
      { location: "/password", messageID: "errors.password-policy.unknown" },
    ]);
  }

  return null;
}
