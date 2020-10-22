import React, { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { LocaleProvider, FormattedMessage } from "@oursky/react-messageformat";
import { ApolloProvider } from "@apollo/client";
import authgear from "@authgear/web";
import Authenticated from "./graphql/portal/Authenticated";
import AppsScreen from "./graphql/portal/AppsScreen";
import CreateAppScreen from "./graphql/portal/CreateAppScreen";
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
import { loadTheme } from "@fluentui/react";

async function loadSystemConfig(): Promise<SystemConfig> {
  const resp = await fetch("/api/system-config.json");
  const config = (await resp.json()) as PartialSystemConfig;
  const mergedConfig = mergeSystemConfig(defaultSystemConfig, config);
  return instantiateSystemConfig(mergedConfig);
}

async function initApp(systemConfig: SystemConfig) {
  loadTheme(systemConfig.themes.main);
  await authgear.configure({
    clientID: systemConfig.authgearClientID,
    endpoint: systemConfig.authgearEndpoint,
  });
}

// ReactAppRoutes defines the routes.
const ReactAppRoutes: React.FC = function ReactAppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="apps/" replace={true} />} />
        <Route path="/apps/" element={<AppsScreen />} />
        <Route path="/apps/create" element={<CreateAppScreen />} />
        <Route path="/app/:appID/*" element={<AppRoot />} />
        <Route path="/oauth-redirect" element={<OAuthRedirect />} />
        <Route
          path="/collaborators/invitation"
          element={<AcceptAdminInvitationScreen />}
        />
      </Routes>
    </BrowserRouter>
  );
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
      <LocaleProvider locale="en" messageByID={MESSAGES}>
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
    <LocaleProvider locale="en" messageByID={systemConfig.translations.en}>
      <ApolloProvider client={client}>
        <SystemConfigContext.Provider value={systemConfig}>
          <div className={styles.root}>
            <Authenticated>
              <ReactAppRoutes />
            </Authenticated>
          </div>
        </SystemConfigContext.Provider>
      </ApolloProvider>
    </LocaleProvider>
  );
};

export default ReactApp;
