import React, { useEffect } from "react";
import authgear from "@authgear/web";
import { useNavigate } from "react-router-dom";

function decodeOAuthState(oauthState: string): Record<string, unknown> {
  return JSON.parse(atob(oauthState));
}

function isString(value: unknown): value is string {
  return typeof value === "string";
}

const OAuthRedirect: React.FC = function OAuthRedirect() {
  const navigate = useNavigate();

  useEffect(() => {
    authgear
      .finishAuthorization()
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
          state: {
            isOAuthRedirect: true,
          },
        });
      })
      .catch((err) => {
        console.error(err);
      });
  }, [navigate]);

  return null;
};

export default OAuthRedirect;
