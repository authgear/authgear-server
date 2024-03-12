import { LocalValidationError } from "../error/validation";

export interface RoleInput {
  key: string;
  name: string;
  description?: string | null;
}

// Ref: pkg/lib/rolesgroups/key.go
const KEY_REGEX = /^[a-zA-Z_][a-zA-Z0-9:_]*$/;
const MAX_KEY_LENGTH = 40;

export function validateRole(
  rawInput: RoleInput
): [RoleInput, LocalValidationError[]] {
  const input = sanitizeRole(rawInput);
  const errors: LocalValidationError[] = [];
  if (!input.key) {
    errors.push({
      location: "/key",
      messageID: "errors.validation.required",
    });
  } else if (!KEY_REGEX.test(input.key)) {
    errors.push({
      location: "/key",
      messageID: "errors.roles.key.validation.format",
    });
  } else if (input.key.length > MAX_KEY_LENGTH) {
    errors.push({
      location: "/key",
      messageID: "errors.validation.maxLength",
      arguments: { expected: MAX_KEY_LENGTH },
    });
  }

  if (!input.name) {
    errors.push({
      location: "/name",
      messageID: "errors.validation.required",
    });
  }
  return [input, errors];
}

export function sanitizeRole(input: RoleInput): RoleInput {
  const description = input.description?.trim();
  return {
    key: input.key.trim(),
    name: input.name.trim(),
    description: description ? description : null,
  };
}
