import { PortalAPIAppConfig } from "../types";

export function canCreateLoginIDIdentity(
  appConfig: PortalAPIAppConfig | null
): boolean {
  const identities = appConfig?.authentication?.identities ?? [];
  // require login ID in identities to create login ID
  return identities.includes("login_id");
}
