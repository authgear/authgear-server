import { useMemo } from "react";
import { OAuthClientConfig } from "../types";

interface Endpoints {
  openidConfiguration: string;
  authorize: string | null;
  token: string;
  userinfo: string | null;
  endSession: string | null;
  jwksUri: string;
}

export function useEndpoints(
  publicOrigin: string,
  applicationType: OAuthClientConfig["x_application_type"]
): Endpoints {
  const endpoints = useMemo(() => {
    switch (applicationType) {
      case "m2m":
        return {
          openidConfiguration: `${publicOrigin}/.well-known/openid-configuration`,
          token: `${publicOrigin}/oauth2/token`,
          jwksUri: `${publicOrigin}/oauth2/jwks`,
          authorize: null,
          userinfo: null,
          endSession: null,
        };
      default:
        return {
          openidConfiguration: `${publicOrigin}/.well-known/openid-configuration`,
          authorize: `${publicOrigin}/oauth2/authorize`,
          token: `${publicOrigin}/oauth2/token`,
          userinfo: `${publicOrigin}/oauth2/userinfo`,
          endSession: `${publicOrigin}/oauth2/end_session`,
          jwksUri: `${publicOrigin}/oauth2/jwks`,
        };
    }
  }, [applicationType, publicOrigin]);

  return endpoints;
}
