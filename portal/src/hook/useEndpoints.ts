import { useMemo } from "react";

interface Endpoints {
  openidConfiguration: string;
  authorize: string;
  token: string;
  userinfo: string;
  endSession: string;
  jwksUri: string;
}

export function useEndpoints(publicOrigin: string): Endpoints {
  const endpoints = useMemo(() => {
    return {
      openidConfiguration: `${publicOrigin}/.well-known/openid-configuration`,
      authorize: `${publicOrigin}/oauth2/authorize`,
      token: `${publicOrigin}/oauth2/token`,
      userinfo: `${publicOrigin}/oauth2/userinfo`,
      endSession: `${publicOrigin}/oauth2/end_session`,
      jwksUri: `${publicOrigin}/oauth2/jwks`,
    };
  }, [publicOrigin]);

  return endpoints;
}
