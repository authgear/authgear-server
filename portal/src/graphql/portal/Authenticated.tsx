import React, {
  useEffect,
  useCallback,
  useMemo,
  useState,
  useContext,
  createContext,
} from "react";
import authgear, {
  PromptOption,
  WebContainer,
  SessionStateChangeReason,
  SessionState,
  AuthenticateResult,
  ConfigureOptions,
} from "@authgear/web";
import { useNavigate, createPath } from "react-router-dom";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useViewerQuery } from "./query/viewerQuery";
import { InternalRedirectState } from "../../InternalRedirect";
import { useReset } from "../../gtm_v2";

interface AuthenticatedContextValue {
  loading: boolean;
  error: unknown;
  authenticated: boolean;
  refetch: () => Promise<unknown>;
}

const DEFAULT_VALUE: AuthenticatedContextValue = {
  loading: true,
  error: null,
  authenticated: false,
  refetch: async () => {},
};

const AuthenticatedContext = createContext(DEFAULT_VALUE);

interface ShowQueryResultProps {
  isAuthenticated: boolean;
  children?: React.ReactElement;
}

function encodeOAuthState(state: Record<string, unknown>): string {
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
  const { loading, error, authenticated, refetch } =
    useContext(AuthenticatedContext);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return <ShowQueryResult isAuthenticated={authenticated} {...ownProps} />;
};

// eslint-disable-next-line @typescript-eslint/no-unnecessary-type-parameters
export async function startReauthentication<S>(
  navigate: ReturnType<typeof useNavigate>,
  state?: S
): Promise<void> {
  const originalPath = createPath(window.location);

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

export function useStartReauthentication<S>(): {
  startReauthentication: typeof startReauthentication<S>;
  isRevealing: boolean;
} {
  const [isRevealing, setIsRevealing] = useState(false);
  const startReauthenticationWithLoading = useCallback(
    async (navigate: ReturnType<typeof useNavigate>, state?: S) => {
      setIsRevealing(true);
      return startReauthentication(navigate, state);
    },
    []
  );
  return {
    startReauthentication: startReauthenticationWithLoading,
    isRevealing,
  };
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

// useFinishAuthentication was introduced to avoid a possible race condition.
// The ultimate source of truth to determine whether the user has authenticated or not is by checking viewer != null.
// Therefore, when we finish authentication, we always refetch the query.
//
// Not doing this will result in a situation where
// 1. authgear.accessToken != null (In the end-user's point of view, he just authenticated, and being redirected back to the portal)
// 2. { loading = false, viewer = null } (because refetch() DOES NOT set loading to true immediately)
// 3. Then Authenticated will redirect the end-user to authenticate again, which is buggy.
export function useFinishAuthentication(): () => Promise<AuthenticateResult> {
  const { refetch } = useContext(AuthenticatedContext);
  const finishAuthentication = useCallback(async () => {
    const result = await authgear.finishAuthentication();
    await refetch();
    return result;
  }, [refetch]);
  return finishAuthentication;
}

export interface ConfigureAuthgearOptions {
  clientID: string;
  endpoint: string;
  sessionType: NonNullable<ConfigureOptions["sessionType"]>;
}

export async function configureAuthgear(
  options: ConfigureAuthgearOptions
): Promise<void> {
  // eslint-disable-next-line no-console -- Output the session type to console for easier debugging.
  console.info("authgear: sessionType = %s", options.sessionType);
  await authgear.configure({
    sessionType: options.sessionType,
    clientID: options.clientID,
    endpoint: options.endpoint,
  });
}

export interface AuthenticatedContextProviderProps {
  children?: React.ReactElement;
}

export function AuthenticatedContextProvider(
  props: AuthenticatedContextProviderProps
): React.ReactElement | null {
  const [sessionState, setSessionState] = useState(authgear.sessionState);
  const { viewer, loading, error, refetch } = useViewerQuery();

  const delegate = useMemo(() => {
    return {
      onSessionStateChange: (
        container: WebContainer,
        _reason: SessionStateChangeReason
      ) => {
        setSessionState(container.sessionState);
        refetch();
      },
    };
  }, [refetch]);

  // Set delegate
  useEffect(() => {
    authgear.delegate = delegate;
  }, [delegate]);

  const value = useMemo(() => {
    let authenticated = false;
    switch (authgear.sessionType) {
      case "cookie":
        // FIXME(authgear-sdk): Update to the version that includes https://github.com/authgear/authgear-sdk-js/pull/336
        // When switching from refresh_token to cookie, with the fix in https://github.com/authgear/authgear-sdk-js/pull/336,
        // authgear SDK will not load refresh token, then thus authgear.fetch will not include
        // Authorization header.
        //
        // We just need to check if we can actually fetch viewer.
        authenticated = viewer != null;
        break;
      case "refresh_token":
        // When switching from cookie to refresh_token, the cookie may left in the browser.
        // So we have to check if authgear SDK does have a stored refresh_token.
        // This checking is reflected by authgear.sessionState.
        authenticated =
          sessionState === SessionState.Authenticated && viewer != null;
        break;

      // So now, switching between cookie and refresh_token is possible and work seamlessly.
      // Of course, the switch implies the end-user has to authenticate again.
    }

    return {
      loading,
      error,
      authenticated,
      refetch,
    };
  }, [sessionState, loading, error, viewer, refetch]);

  return (
    <AuthenticatedContext.Provider value={value}>
      {props.children}
    </AuthenticatedContext.Provider>
  );
}

export default Authenticated;
