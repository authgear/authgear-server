import React, { useEffect, useCallback } from "react";
import authgear, { PromptOption } from "@authgear/web";
import { useNavigate } from "react-router-dom";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useViewerQuery } from "./query/viewerQuery";
import { InternalRedirectState } from "../../InternalRedirect";
import { useReset } from "../../gtm_v2";

interface ShowQueryResultProps {
  isAuthenticated: boolean;
  children?: React.ReactElement;
}

function encodeOAuthState(state: Record<string, unknown>): string {
  // eslint-disable-next-line @typescript-eslint/no-deprecated
  return btoa(JSON.stringify(state));
}

const ShowQueryResult: React.VFC<ShowQueryResultProps> =
  function ShowQueryResult(props: ShowQueryResultProps) {
    const { isAuthenticated } = props;

    const redirectURI = window.location.origin + "/oauth-redirect";
    const originalPath = `${window.location.pathname}${window.location.search}`;

    useEffect(() => {
      if (!isAuthenticated) {
        // Normally we should call endAuthorization after being redirected back to here.
        // But we know that we are first party app and are using response_type=none so
        // we can skip that.
        authgear
          .startAuthentication({
            redirectURI,
            prompt: PromptOption.Login,
            state: encodeOAuthState({
              originalPath,
            }),
          })
          .catch((err) => {
            console.error(err);
          });
      }
    }, [isAuthenticated, redirectURI, originalPath]);

    if (isAuthenticated) {
      return props.children ?? null;
    }

    return null;
  };

interface Props {
  children?: React.ReactElement;
}

// CAVEAT: <Authenticated><Route path="/foobar/:id"/></Authenticated> will cause useParams to return empty object :(
const Authenticated: React.VFC<Props> = function Authenticated(
  ownProps: Props
) {
  const { loading, error, viewer, refetch } = useViewerQuery();

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return <ShowQueryResult isAuthenticated={viewer != null} {...ownProps} />;
};

// eslint-disable-next-line @typescript-eslint/no-unnecessary-type-parameters
export async function startReauthentication<S>(
  navigate: ReturnType<typeof useNavigate>,
  state?: S
): Promise<void> {
  const originalPath = `${window.location.pathname}${window.location.search}`;

  await authgear.refreshIDToken();
  // If the user cannot reauthenticate, we perform a internal-redirect
  // to emulate the effect of redirection after reauthentication.
  if (!authgear.canReauthenticate()) {
    navigate("/internal-redirect", {
      state: {
        originalPath,
        state,
      } as InternalRedirectState,
      replace: true,
    });
    return;
  }

  const redirectURI = window.location.origin + "/oauth-redirect";
  await authgear.startReauthentication({
    redirectURI,
    state: encodeOAuthState({
      originalPath,
      state,
    }),
  });
}

export function useLogout(): () => Promise<void> {
  const redirectURI = window.location.origin + "/";
  const reset = useReset();
  const logout = useCallback(async () => {
    await authgear.logout({
      redirectURI,
    });
    reset();
  }, [redirectURI, reset]);
  return logout;
}

export default Authenticated;
