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
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import {
  LocaleProvider,
  FormattedMessage,
  Context,
} from "@oursky/react-messageformat";
import { ApolloProvider } from "@apollo/client";
import { Helmet, HelmetProvider } from "react-helmet-async";
import AppRoot from "./AppRoot";
import MESSAGES from "./locale-data/en.json";
import styles from "./ReactApp.module.css";
import { SystemConfigContext } from "./context/SystemConfigContext";
import {
  SystemConfig,
  PartialSystemConfig,
  defaultSystemConfig,
  instantiateSystemConfig,
  mergeSystemConfig,
} from "./system-config";
import { loadTheme, ILinkProps } from "@fluentui/react";
import ExternalLink from "./ExternalLink";
import Link from "./Link";
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

const AppsScreen = lazy(async () => import("./graphql/portal/AppsScreen"));
const CreateProjectScreen = lazy(
  async () => import("./graphql/portal/CreateProjectScreen")
);
const OnboardingSurveyScreen = lazy(
  async () => import("./OnboardingSurveyScreen")
);
const ProjectWizardScreen = lazy(
  async () => import("./graphql/portal/ProjectWizardScreen")
);
const OnboardingRedirect = lazy(async () => import("./OnboardingRedirect"));
const OAuthRedirect = lazy(async () => import("./OAuthRedirect"));
const AcceptAdminInvitationScreen = lazy(
  async () => import("./graphql/portal/AcceptAdminInvitationScreen")
);

const StoryBookScreen = lazy(async () => import("./StoryBookScreen"));

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

// ReactAppRoutes defines the routes.
const ReactAppRoutes: React.VFC = function ReactAppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          index={true}
          element={
            <Authenticated>
              <Navigate to="/projects" replace={true} />
            </Authenticated>
          }
        />

        <Route path="/projects">
          <Route
            index={true}
            element={
              <Authenticated>
                <Suspense fallback={<ShowLoading />}>
                  <AppsScreen />
                </Suspense>
              </Authenticated>
            }
          />
          <Route
            path="create"
            element={
              <Authenticated>
                <Suspense fallback={<ShowLoading />}>
                  <CreateProjectScreen />
                </Suspense>
              </Authenticated>
            }
          />
        </Route>
        <Route path="onboarding-survey">
          <Route
            index={true}
            // @ts-expect-error
            path="*"
            element={
              <Authenticated>
                <Suspense fallback={<ShowLoading />}>
                  <OnboardingSurveyScreen />
                </Suspense>
              </Authenticated>
            }
          />
        </Route>
        <Route path="/project">
          <Route index={true} element={<Navigate to="/" />} />
          <Route path=":appID">
            <Route
              index={true}
              // @ts-expect-error
              path="*"
              element={
                <Authenticated>
                  <AppContextProvider>
                    <AppRoot />
                  </AppContextProvider>
                </Authenticated>
              }
            />
            <Route path="wizard">
              <Route
                index={true}
                // @ts-expect-error
                path="*"
                element={
                  <Authenticated>
                    <Suspense fallback={<ShowLoading />}>
                      <AppContextProvider>
                        <ProjectWizardScreen />
                      </AppContextProvider>
                    </Suspense>
                  </Authenticated>
                }
              />
            </Route>
          </Route>
        </Route>

        <Route
          path="/oauth-redirect"
          element={
            <Suspense fallback={<ShowLoading />}>
              <OAuthRedirect />
            </Suspense>
          }
        />

        <Route path="/internal-redirect" element={<InternalRedirect />} />

        <Route
          path="/onboarding-redirect"
          element={
            <Suspense fallback={<ShowLoading />}>
              <OnboardingRedirect />
            </Suspense>
          }
        />

        <Route
          path="/collaborators/invitation"
          element={
            <Suspense fallback={<ShowLoading />}>
              <AcceptAdminInvitationScreen />
            </Suspense>
          }
        />

        <Route
          path="/storybook"
          element={
            <Suspense fallback={<ShowLoading />}>
              <StoryBookScreen />
            </Suspense>
          }
        />
      </Routes>
    </BrowserRouter>
  );
};

const PortalRoot = function PortalRoot() {
  const { renderToString } = useContext(Context);
  return (
    <>
      <Helmet>
        <title>{renderToString("system.title")} </title>
      </Helmet>
      <div className={styles.root}>
        <ReactAppRoutes />
      </div>
    </>
  );
};

const DocLink: React.VFC<ILinkProps> = (props: ILinkProps) => {
  return <ExternalLink {...props} />;
};

const defaultComponents = {
  ExternalLink,
  ReactRouterLink: Link,
  DocLink,
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
      <LocaleProvider
        locale="en"
        messageByID={MESSAGES}
        defaultComponents={defaultComponents}
      >
        <p>
          <FormattedMessage id="error.failed-to-initialize-app" />
        </p>
      </LocaleProvider>
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
            <LocaleProvider
              locale="en"
              messageByID={systemConfig.translations.en}
              defaultComponents={defaultComponents}
            >
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
            </LocaleProvider>
          </LoadingContextProvider>
        </ErrorContextProvider>
      </GTMProvider>
    </ErrorBoundary>
  );
};

export default ReactApp;
