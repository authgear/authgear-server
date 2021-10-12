import { StandardAttributes } from "../types";

// getEndUserAccountIdentifier returns a string that helps end-user to identifier their account.
export function getEndUserAccountIdentifier(
  attrs: StandardAttributes
): string | undefined {
  return attrs.email ?? attrs.preferred_username ?? attrs.phone_number;
}
