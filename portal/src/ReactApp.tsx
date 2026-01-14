import React, {
  useContext,
  useEffect,
  useState,
  Suspense,
  lazy,
  useCallback,
  useMemo,
} from "react";
import {
  Exception as SentryException,
  ErrorEvent as SentryErrorEvent,
  EventHint,
  ErrorBoundary,
  init as sentryInit,
} from "@sentry/react";
import {
  createBrowserRouter,
  RouterProvider,
  Navigate,
  Outlet,
} from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { ApolloProvider } from "@apollo/client";
import { Helmet, HelmetProvider } from "react-helmet-async";
import AppRoot from "./AppRoot";
import styles from "./ReactApp.module.css";
import { SystemConfigContext } from "./context/SystemConfigContext";
import {
  SystemConfig,
  PartialSystemConfig,
  defaultSystemConfig,
  instantiateSystemConfig,
  mergeSystemConfig,
} from "./system-config";
import { loadTheme } from "@fluentui/react";
import Authenticated, {
  configureAuthgear,
  AuthenticatedContextProvider,
} from "./graphql/portal/Authenticated";
import InternalRedirect from "./InternalRedirect";
import { LoadingContextProvider } from "./hook/loading";
import { ErrorContextProvider } from "./hook/error";
import ShowLoading from "./ShowLoading";
import GTMProvider from "./GTMProvider";
import { FallbackComponent } from "./FlavoredErrorBoundSuspense";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";
import { extractRawID } from "./util/graphql";
import { useIdentify } from "./gtm_v2";
import AppContextProvider from "./AppContextProvider";
import {
  PortalClientProvider,
  createCache,
  createClient,
} from "./graphql/portal/apollo";
import { ViewerQueryDocument } from "./graphql/portal/query/viewerQuery.generated";
import { UnauthenticatedDialog } from "./components/auth/UnauthenticatedDialog";
import {
  UnauthenticatedDialogContext,
  UnauthenticatedDialogContextValue,
} from "./components/auth/UnauthenticatedDialogContext";
import { isNetworkError } from "./util/error";
import { ToastProvider } from "./components/v2/Toast/Toast";
import { ThemeProvider } from "./components/v2/ThemeProvider/ThemeProvider";
import { AppLocaleProvider } from "./components/common/AppLocaleProvider";

const AppsScreen = lazy(async () => import("./graphql/portal/AppsScreen"));

const OnboardingSurveyScreen = lazy(
  async () => import("./screens/v2/OnboardingSurvey/OnboardingSurveyScreen")
);
const ProjectWizardScreenV2 = lazy(
  async () => import("./screens/v2/ProjectWizard/ProjectWizardScreen")
);
const OnboardingRedirect = lazy(async () => import("./OnboardingRedirect"));
const OAuthRedirect = lazy(async () => import("./OAuthRedirect"));
const AcceptAdminInvitationScreen = lazy(
  async () => import("./graphql/portal/AcceptAdminInvitationScreen")
);

async function loadSystemConfig(): Promise<SystemConfig> {
  const resp = await fetch("/api/system-config.json");
  const config = (await resp.json()) as PartialSystemConfig;
  const mergedConfig = mergeSystemConfig(defaultSystemConfig, config);
  return instantiateSystemConfig(mergedConfig);
}

function isPosthogResetGroupsException(ex: SentryException) {
  return ex.type === "TypeError" && ex.value?.includes("posthog.resetGroups");
}
function isPosthogResetGroupsEvent(event: SentryErrorEvent) {
  return event.exception?.values?.some(isPosthogResetGroupsException) ?? false;
}

// DEV-1767: Unknown cause on posthog error, silence for now
function sentryBeforeSend(event: SentryErrorEvent, hint: EventHint) {
  if (isPosthogResetGroupsEvent(event)) {
    return null;
  }
  if (isNetworkError(hint.originalException)) {
    console.warn("skip sending network error to Sentry");
    console.warn(hint.originalException);
    return null;
  }
  return event;
}

async function initApp(systemConfig: SystemConfig) {
  if (systemConfig.sentryDSN !== "") {
    sentryInit({
      dsn: systemConfig.sentryDSN,
      tracesSampleRate: 0.0,
      beforeSend: sentryBeforeSend,
    });
  }

  loadTheme(systemConfig.themes.main);
}

const router = createBrowserRouter([
  {
    path: "/",
    element: <Outlet />,
    children: [
      {
        index: true,
        element: (
          <Authenticated>
            <Navigate to="/projects" replace={true} />
          </Authenticated>
        ),
      },
      {
        path: "projects",
        children: [
          {
            index: true,
            element: (
              <Authenticated>
                <Suspense fallback={<ShowLoading />}>
                  <AppsScreen />
                </Suspense>
              </Authenticated>
            ),
          },
          {
            path: "create",
            element: (
              <Authenticated>
                <Suspense fallback={<ShowLoading />}>
                  <ProjectWizardScreenV2 />
                </Suspense>
              </Authenticated>
            ),
          },
        ],
      },
      {
        path: "onboarding-survey",
        children: [
          {
            index: true,
            path: "*",
            element: (
              <Authenticated>
                <Suspense fallback={<ShowLoading />}>
                  <OnboardingSurveyScreen />
                </Suspense>
              </Authenticated>
            ),
          },
        ],
      },
      {
        path: "project",
        element: <Outlet />,
        children: [
          {
            index: true,
            element: <Navigate to="/" />,
          },
          {
            path: ":appID",
            element: <Outlet />,
            children: [
              {
                index: true,
                path: "*",
                element: (
                  <Authenticated>
                    <AppContextProvider>
                      <AppRoot />
                    </AppContextProvider>
                  </Authenticated>
                ),
              },
              {
                path: "wizard",
                children: [
                  {
                    index: true,
                    path: "*",
                    element: (
                      <Authenticated>
                        <Suspense fallback={<ShowLoading />}>
                          <AppContextProvider>
                            <ProjectWizardScreenV2 />
                          </AppContextProvider>
                        </Suspense>
                      </Authenticated>
                    ),
                  },
                ],
              },
            ],
          },
        ],
      },
      {
        path: "oauth-redirect",
        element: (
          <Suspense fallback={<ShowLoading />}>
            <OAuthRedirect />
          </Suspense>
        ),
      },
      {
        path: "internal-redirect",
        element: <InternalRedirect />,
      },
      {
        path: "onboarding-redirect",
        element: (
          <Suspense fallback={<ShowLoading />}>
            <OnboardingRedirect />
          </Suspense>
        ),
      },
      {
        path: "collaborators/invitation",
        element: (
          <Suspense fallback={<ShowLoading />}>
            <AcceptAdminInvitationScreen />
          </Suspense>
        ),
      },
    ],
  },
]);

const PortalRoot = function PortalRoot() {
  const { renderToString } = useContext(Context);
  return (
    <>
      <Helmet>
        <title>{renderToString("system.title")} </title>
      </Helmet>
      <ThemeProvider>
        <ToastProvider>
          <div className={styles.root}>
            <RouterProvider router={router} />
          </div>
        </ToastProvider>
      </ThemeProvider>
    </>
  );
};

export interface LoadCurrentUserProps {
  children?: React.ReactNode;
}

const LoadCurrentUser: React.VFC<LoadCurrentUserProps> =
  function LoadCurrentUser({ children }: LoadCurrentUserProps) {
    const { loading, viewer } = useViewerQuery();

    const identify = useIdentify();
    useEffect(() => {
      if (viewer) {
        const userID = extractRawID(viewer.id);
        const email = viewer.email ?? undefined;

        identify(userID, email);
      }
    }, [viewer, identify]);

    if (loading) {
      return (
        <div className={styles.root}>
          <ShowLoading />
        </div>
      );
    }

    return <>{children}</>;
  };

// ReactApp is responsible for fetching runtime config and initialize authgear SDK.
const ReactApp: React.VFC = function ReactApp() {
  const [systemConfig, setSystemConfig] = useState<SystemConfig | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [displayUnauthenticatedDialog, setDisplayUnauthenticatedDialog] =
    useState(false);

  const [apolloClient] = useState(() => {
    const cache = createCache();
    return createClient({
      cache: cache,
      onLogout: () => {
        setDisplayUnauthenticatedDialog(true);
      },
    });
  });

  const onUnauthenticatedDialogConfirm = useCallback(() => {
    apolloClient.cache.writeQuery({
      query: ViewerQueryDocument,
      data: {
        viewer: null,
      },
    });
  }, [apolloClient.cache]);

  const unauthenticatedDialogContextValue =
    useMemo<UnauthenticatedDialogContextValue>(() => {
      return {
        setDisplayUnauthenticatedDialog,
      };
    }, []);

  useEffect(() => {
    if (!systemConfig && error == null) {
      loadSystemConfig()
        .then(async (cfg) => {
          await initApp(cfg);
          await configureAuthgear({
            clientID: cfg.authgearClientID,
            endpoint: cfg.authgearEndpoint,
            sessionType: cfg.authgearWebSDKSessionType,
          });
          setSystemConfig(cfg);
        })
        .catch((err) => {
          setError(err);
        });
    }
  }, [systemConfig, error]);

  if (error != null) {
    return (
      <AppLocaleProvider>
        <p>
          <FormattedMessage id="error.failed-to-initialize-app" />
        </p>
      </AppLocaleProvider>
    );
  } else if (!systemConfig) {
    // Avoid rendering components from @fluentui/react, since themes are not loaded yet.
    return null;
  }

  return (
    <ErrorBoundary fallback={FallbackComponent}>
      <GTMProvider containerID={systemConfig.gtmContainerID}>
        <ErrorContextProvider>
          <LoadingContextProvider>
            <AppLocaleProvider systemConfig={systemConfig}>
              <HelmetProvider>
                <PortalClientProvider value={apolloClient}>
                  <ApolloProvider client={apolloClient}>
                    <SystemConfigContext.Provider value={systemConfig}>
                      <AuthenticatedContextProvider>
                        <LoadCurrentUser>
                          <UnauthenticatedDialogContext.Provider
                            value={unauthenticatedDialogContextValue}
                          >
                            <PortalRoot />
                          </UnauthenticatedDialogContext.Provider>
                          <UnauthenticatedDialog
                            isHidden={!displayUnauthenticatedDialog}
                            onConfirm={onUnauthenticatedDialogConfirm}
                          />
                        </LoadCurrentUser>
                      </AuthenticatedContextProvider>
                    </SystemConfigContext.Provider>
                  </ApolloProvider>
                </PortalClientProvider>
              </HelmetProvider>
            </AppLocaleProvider>
          </LoadingContextProvider>
        </ErrorContextProvider>
      </GTMProvider>
    </ErrorBoundary>
  );
};

export default ReactApp;
