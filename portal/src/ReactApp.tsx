import React, { useContext, useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import {
  LocaleProvider,
  FormattedMessage,
  Context,
} from "@oursky/react-messageformat";
import { ApolloProvider } from "@apollo/client";
import authgear from "@authgear/web";
import { Helmet, HelmetProvider } from "react-helmet-async";
import AppsScreen from "./graphql/portal/AppsScreen";
import CreateProjectScreen from "./graphql/portal/CreateProjectScreen";
import ProjectWizardScreen from "./graphql/portal/ProjectWizardScreen";
import AppRoot from "./AppRoot";
import MESSAGES from "./locale-data/en.json";
import { client } from "./graphql/portal/apollo";
import { registerLocale } from "i18n-iso-countries";
import i18nISOCountriesEnLocale from "i18n-iso-countries/langs/en.json";
import styles from "./ReactApp.module.scss";
import OAuthRedirect from "./OAuthRedirect";
import AcceptAdminInvitationScreen from "./graphql/portal/AcceptAdminInvitationScreen";
import { SystemConfigContext } from "./context/SystemConfigContext";
import {
  SystemConfig,
  PartialSystemConfig,
  defaultSystemConfig,
  instantiateSystemConfig,
  mergeSystemConfig,
} from "./system-config";
import { loadTheme, Link as FluentLink, ILinkProps } from "@fluentui/react";
import ProjectWizardDoneScreen from "./graphql/portal/ProjectWizardDoneScreen";
import OnboardingRedirect from "./OnboardingRedirect";
import { ReactRouterLink, ReactRouterLinkProps } from "./ReactRouterLink";
import Authenticated from "./graphql/portal/Authenticated";

async function loadSystemConfig(): Promise<SystemConfig> {
  const resp = await fetch("/api/system-config.json");
  const config = (await resp.json()) as PartialSystemConfig;
  const mergedConfig = mergeSystemConfig(defaultSystemConfig, config);
  return instantiateSystemConfig(mergedConfig);
}

async function initApp(systemConfig: SystemConfig) {
  loadTheme(systemConfig.themes.main);
  await authgear.configure({
    sessionType: "cookie",
    clientID: systemConfig.authgearClientID,
    endpoint: systemConfig.authgearEndpoint,
  });
}

// ReactAppRoutes defines the routes.
const ReactAppRoutes: React.FC = function ReactAppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route
          path="/"
          element={
            <Authenticated>
              <Navigate to="projects/" replace={true} />
            </Authenticated>
          }
        />
        <Route
          path="/projects/"
          element={
            <Authenticated>
              <AppsScreen />
            </Authenticated>
          }
        />
        <Route
          path="/projects/create"
          element={
            <Authenticated>
              <CreateProjectScreen />
            </Authenticated>
          }
        />
        <Route
          path="/project/:appID/*"
          element={
            <Authenticated>
              <AppRoot />
            </Authenticated>
          }
        />
        <Route
          path="/project/:appID/wizard/*"
          element={
            <Authenticated>
              <ProjectWizardScreen />
            </Authenticated>
          }
        />
        <Route
          path="/project/:appID/wizard/done"
          element={
            <Authenticated>
              <ProjectWizardDoneScreen />
            </Authenticated>
          }
        />
        <Route path="/oauth-redirect" element={<OAuthRedirect />} />
        <Route path="/onboarding-redirect" element={<OnboardingRedirect />} />
        <Route
          path="/"
          element={
            <Authenticated>
              <Navigate to="projects/" replace={true} />
            </Authenticated>
          }
        />
        <Route
          path="/collaborators/invitation"
          element={<AcceptAdminInvitationScreen />}
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

const PortalLink = React.forwardRef<HTMLAnchorElement, ReactRouterLinkProps>(
  function LinkWithRef({ ...rest }, ref) {
    return <ReactRouterLink {...rest} ref={ref} component={FluentLink} />;
  }
);

function ExternalLink(props: ILinkProps) {
  return <FluentLink target="_blank" rel="noreferrer" {...props} />;
}

const defaultComponents = {
  ExternalLink,
  ReactRouterLink: PortalLink,
};

// ReactApp is responsible for fetching runtime config and initialize authgear SDK.
const ReactApp: React.FC = function ReactApp() {
  const [systemConfig, setSystemConfig] = useState<SystemConfig | null>(null);
  const [error, setError] = useState<null | unknown>(null);

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

  // register locale for country code translation
  registerLocale(i18nISOCountriesEnLocale);

  return (
    <LocaleProvider
      locale="en"
      messageByID={systemConfig.translations.en}
      defaultComponents={defaultComponents}
    >
      <HelmetProvider>
        <ApolloProvider client={client}>
          <SystemConfigContext.Provider value={systemConfig}>
            <PortalRoot />
          </SystemConfigContext.Provider>
        </ApolloProvider>
      </HelmetProvider>
    </LocaleProvider>
  );
};

export default ReactApp;
