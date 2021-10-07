import { Identity } from "../types";

export interface UserInfo {
  username: string | null;
  phone: string | null;
  email: string | null;
}

export function extractUserInfoFromIdentities(
  identities: Identity[]
): UserInfo {
  const claimsList = identities.map((identity) => identity.claims);

  const email =
    claimsList.map((claims) => claims.email).filter(Boolean)[0] ?? null;
  const username =
    claimsList.map((claims) => claims.preferred_username).filter(Boolean)[0] ??
    null;
  const phone =
    claimsList.map((claims) => claims.phone_number).filter(Boolean)[0] ?? null;

  return { email, username, phone };
}
