import { LocalValidationError } from "../error/validation";
import { Role } from "../graphql/adminapi/globalTypes.generated";
import { processSearchKeyword } from "../util/search";

export interface CreatableRole
  extends Pick<Role, "key" | "name" | "description"> {}

// Ref: pkg/lib/rolesgroups/key.go
const KEY_REGEX = /^[a-zA-Z_][a-zA-Z0-9:_]*$/;
const MAX_KEY_LENGTH = 40;

export function validateRole(
  rawInput: CreatableRole
): [CreatableRole, LocalValidationError[]] {
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

export function sanitizeRole(input: CreatableRole): CreatableRole {
  const description = input.description?.trim();
  const name = input.name?.trim();
  return {
    key: input.key.trim(),
    name: name ? name : null,
    description: description ? description : null,
  };
}

export interface SearchableRole extends Pick<Role, "id" | "key" | "name"> {}

export function searchRoles<R extends SearchableRole>(
  roles: R[],
  searchKeyword: string
): R[] {
  if (searchKeyword === "") {
    return roles;
  }
  const keywords = processSearchKeyword(searchKeyword);
  return roles.filter((role) => {
    const roleID = role.id.toLowerCase();
    const roleKey = role.key.toLowerCase();
    const roleName = role.name?.toLowerCase();
    return keywords.every((keyword) => {
      if (roleID === keyword) {
        return true;
      }
      if (roleKey.includes(keyword)) {
        return true;
      }
      if (roleName?.includes(keyword)) {
        return true;
      }
      return false;
    });
  });
}
