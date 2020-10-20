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
  RuntimeConfig,
  RuntimeConfigContext,
} from "./context/RuntimeConfigContext";

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
  const [runtimeConfigState, setRuntimeConfig] = useState<RuntimeConfig | null>(
    null
  );

  useEffect(() => {
    if (!configured && !loading && error == null) {
      setLoading(true);

      fetch("/api/runtime-config.json")
        .then(async (response) => {
          const runtimeConfig = await response.json();
          setRuntimeConfig(runtimeConfig);
          await authgear.configure({
            clientID: runtimeConfig.authgear_client_id,
            endpoint: runtimeConfig.authgear_endpoint,
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
        <RuntimeConfigContext.Provider value={runtimeConfigState}>
          <div className={styles.root}>{children}</div>
        </RuntimeConfigContext.Provider>
      </ApolloProvider>
    </LocaleProvider>
  );
};

export default ReactApp;
