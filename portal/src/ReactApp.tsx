import React, {
  useContext,
  useEffect,
  useState,
  Suspense,
  lazy,
  useMemo,
} from "react";
import { init as sentryInit } from "@sentry/react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import {
  LocaleProvider,
  FormattedMessage,
  Context,
} from "@oursky/react-messageformat";
import { ApolloProvider } from "@apollo/client";
import authgear from "@authgear/web";
import { Helmet, HelmetProvider } from "react-helmet-async";
import AppRoot from "./AppRoot";
import MESSAGES from "./locale-data/en.json";
import { client } from "./graphql/portal/apollo";
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
import Authenticated from "./graphql/portal/Authenticated";
import InternalRedirect from "./InternalRedirect";
import { LoadingContextProvider } from "./hook/loading";
import { ErrorContextProvider } from "./hook/error";
import ShowLoading from "./ShowLoading";
import GTMProvider, {
  AuthgearGTMEvent,
  AuthgearGTMEventType,
  useMakeAuthgearGTMEventDataAttributes,
  useGTMDispatch,
  useAuthgearGTMEventBase,
} from "./GTMProvider";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";
import { extractRawID } from "./util/graphql";

const AppsScreen = lazy(async () => import("./graphql/portal/AppsScreen"));
const CreateProjectScreen = lazy(
  async () => import("./graphql/portal/CreateProjectScreen")
);
const ProjectWizardScreen = lazy(
  async () => import("./graphql/portal/ProjectWizardScreen")
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

async function initApp(systemConfig: SystemConfig) {
  if (systemConfig.sentryDSN !== "") {
    sentryInit({
      dsn: systemConfig.sentryDSN,
      tracesSampleRate: 0.0,
    });
  }

  loadTheme(systemConfig.themes.main);
  await authgear.configure({
    sessionType: "cookie",
    clientID: systemConfig.authgearClientID,
    endpoint: systemConfig.authgearEndpoint,
  });
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

        <Route path="/project">
          <Route path=":appID">
            <Route
              index={true}
              // @ts-expect-error
              path="*"
              element={
                <Authenticated>
                  <AppRoot />
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
                      <ProjectWizardScreen />
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
  const makeGTMEventDataAttributes = useMakeAuthgearGTMEventDataAttributes();
  const gtmEventDataAttributes = useMemo(() => {
    return makeGTMEventDataAttributes({
      event: AuthgearGTMEventType.ClickedDocLink,
      eventDataAttributes: {
        "doc-link": props.href ?? "",
      },
    });
  }, [makeGTMEventDataAttributes, props.href]);

  return <ExternalLink {...gtmEventDataAttributes} {...props} />;
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

    const gtmEvent = useAuthgearGTMEventBase();
    const sendDataToGTM = useGTMDispatch();
    useEffect(() => {
      if (viewer) {
        const event: AuthgearGTMEvent = {
          ...gtmEvent,
          event: AuthgearGTMEventType.Identified,
          event_data: {
            user_id: extractRawID(viewer.id),
            email: viewer.email ?? undefined,
          },
        };
        sendDataToGTM(event);
      }
    }, [viewer, gtmEvent, sendDataToGTM]);

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

  useEffect(() => {
    if (!systemConfig && error == null) {
      loadSystemConfig()
        .then(async (cfg) => {
          await initApp(cfg);
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
    <GTMProvider containerID={systemConfig.gtmContainerID}>
      <ErrorContextProvider>
        <LoadingContextProvider>
          <LocaleProvider
            locale="en"
            messageByID={systemConfig.translations.en}
            defaultComponents={defaultComponents}
          >
            <HelmetProvider>
              <ApolloProvider client={client}>
                <SystemConfigContext.Provider value={systemConfig}>
                  <LoadCurrentUser>
                    <PortalRoot />
                  </LoadCurrentUser>
                </SystemConfigContext.Provider>
              </ApolloProvider>
            </HelmetProvider>
          </LocaleProvider>
        </LoadingContextProvider>
      </ErrorContextProvider>
    </GTMProvider>
  );
};

export default ReactApp;
