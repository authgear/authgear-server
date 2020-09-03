export interface UserInfo {
  username: string | null;
  phone: string | null;
  email: string | null;
}

export interface IdentityClaims extends GQL_JSONObject {
  email?: string;
  preferred_username?: string;
  phone_number?: string;
}

interface Identity {
  id: string;
  claims: IdentityClaims;
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
