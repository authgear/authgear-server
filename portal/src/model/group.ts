import { LocalValidationError } from "../error/validation";
import { Group } from "../graphql/adminapi/globalTypes.generated";
import { processSearchKeyword } from "../util/search";

export interface CreatableGroup
  extends Pick<Group, "key" | "name" | "description"> {}

// Ref: pkg/lib/rolesgroups/key.go
const KEY_REGEX = /^[a-zA-Z_][a-zA-Z0-9:_]*$/;
const MAX_KEY_LENGTH = 40;

export function generateGroupKeyFromName(name: string): string {
  let processedName = name.trim();
  // Replace all whitespace, -, + with _
  processedName = processedName.replace(/[\s\-+]/g, "_");
  // Remove all unallowed characters
  processedName = processedName.replace(/[^a-zA-Z0-9:_]/g, "");
  // Remove all unallowed prefix characters
  processedName = processedName.replace(/^[^a-zA-Z_]*/, "");
  return processedName;
}

export function validateGroup(
  rawInput: CreatableGroup
): [CreatableGroup, LocalValidationError[]] {
  const input = sanitizeGroup(rawInput);
  const errors: LocalValidationError[] = [];
  if (!input.key) {
    errors.push({
      location: "/key",
      messageID: "errors.validation.required",
    });
  } else if (!KEY_REGEX.test(input.key)) {
    errors.push({
      location: "/key",
      messageID: "errors.groups.key.validation.format",
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

export function sanitizeGroup(input: CreatableGroup): CreatableGroup {
  const key = input.key.trim();
  const description = input.description?.trim();
  const name = input.name?.trim();
  return {
    key: key ? key : generateGroupKeyFromName(name ?? ""),
    name: name ? name : null,
    description: description ? description : null,
  };
}

export interface SearchableGroup extends Pick<Group, "id" | "key" | "name"> {}

export function searchGroups<G extends SearchableGroup>(
  groups: G[],
  searchKeyword: string
): G[] {
  if (searchKeyword === "") {
    return groups;
  }
  const keywords = processSearchKeyword(searchKeyword);
  return groups.filter((group) => {
    const groupID = group.id.toLowerCase();
    const groupKey = group.key.toLowerCase();
    const groupName = group.name?.toLowerCase();

    return keywords.every((keyword) => {
      if (groupID === keyword) {
        return true;
      }
      if (groupKey.includes(keyword)) {
        return true;
      }
      if (groupName?.includes(keyword)) {
        return true;
      }
      return false;
    });
  });
}
