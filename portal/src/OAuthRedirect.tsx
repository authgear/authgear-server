import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useFinishAuthentication } from "./graphql/portal/Authenticated";

function decodeOAuthState(oauthState: string): Record<string, unknown> {
  // eslint-disable-next-line @typescript-eslint/no-deprecated
  return JSON.parse(atob(oauthState));
}

function isString(value: unknown): value is string {
  return typeof value === "string";
}

const OAuthRedirect: React.VFC = function OAuthRedirect() {
  const navigate = useNavigate();
  const finishAuthentication = useFinishAuthentication();

  useEffect(() => {
    finishAuthentication()
      .then((result) => {
        const state = result.state ? decodeOAuthState(result.state) : null;
        let navigateToPath = "/";
        if (state && isString(state.originalPath)) {
          navigateToPath = state.originalPath;
        }
        // if user start the authorization from "/"
        // redirect to /onboarding-redirect
        // OnboardingRedirect will further redirect user to create app
        // if the user doesn't have any apps
        if (navigateToPath === "/") {
          navigateToPath = "/onboarding-redirect";
        }
        navigate(navigateToPath, {
          state: state?.state as any,
        });
      })
      .catch((err) => {
        console.error(err);
      });
  }, [navigate, finishAuthentication]);

  return null;
};

export default OAuthRedirect;
