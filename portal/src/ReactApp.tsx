import React, { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { LocaleProvider, FormattedMessage } from "@oursky/react-messageformat";
import { ApolloProvider } from "@apollo/client";
import authgear from "@authgear/web";
import ShowLoading from "./ShowLoading";
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
import {
  SystemConfig,
  SystemConfigContext,
} from "./context/SystemConfigContext";

async function loadSystemConfig(): Promise<SystemConfig> {
  const resp = await fetch("/api/system-config.json");
  return resp.json();
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
  const [configured, setConfigured] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<null | unknown>(null);
  const [systemConfig, setSystemConfig] = useState<SystemConfig | null>(null);

  useEffect(() => {
    if (!configured && !loading && error == null) {
      setLoading(true);

      loadSystemConfig()
        .then(async (cfg) => {
          setSystemConfig(cfg);
          await authgear.configure({
            clientID: cfg.authgearClientID,
            endpoint: cfg.authgearEndpoint,
          });
          setConfigured(true);
        })
        .then(
          () => {
            setLoading(false);
          },
          (err) => {
            setLoading(false);
            setError(err);
          }
        );
    }
  }, [configured, loading, error]);

  let children: React.ReactElement;

  if (error != null) {
    children = (
      <p>
        <FormattedMessage id="error.failed-to-initialize-app" />
      </p>
    );
  } else if (loading) {
    children = <ShowLoading />;
  } else {
    children = (
      <Authenticated>
        <ReactAppRoutes />
      </Authenticated>
    );
  }

  // register locale for country code translation
  registerLocale(i18nISOCountriesEnLocale);

  return (
    <LocaleProvider locale="en" messageByID={MESSAGES}>
      <ApolloProvider client={client}>
        <SystemConfigContext.Provider value={systemConfig}>
          <div className={styles.root}>{children}</div>
        </SystemConfigContext.Provider>
      </ApolloProvider>
    </LocaleProvider>
  );
};

export default ReactApp;
