import React, { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { LocaleProvider, FormattedMessage } from "@oursky/react-messageformat";
import { ApolloProvider } from "@apollo/client";
import authgear from "@authgear/web";
import ShowLoading from "./ShowLoading";
import Authenticated from "./graphql/portal/Authenticated";
import AppsScreen from "./graphql/portal/AppsScreen";
import AppRoot from "./AppRoot";
import MESSAGES from "./locale-data/en.json";
import { client } from "./graphql/portal/apollo";
import styles from "./ReactApp.module.scss";

// ReactAppRoutes defines the routes.
const ReactAppRoutes: React.FC = function ReactAppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="apps/" replace={true} />} />
        <Route path="/apps/" element={<AppsScreen />} />
        <Route path="/apps/:appID/*" element={<AppRoot />} />
      </Routes>
    </BrowserRouter>
  );
};

// ReactApp is responsible for fetching runtime config and initialize authgear SDK.
const ReactApp: React.FC = function ReactApp() {
  const [configured, setConfigured] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<null | unknown>(null);

  useEffect(() => {
    if (!configured && !loading && error == null) {
      setLoading(true);

      fetch("/api/runtime-config.json")
        .then(async (response) => {
          const runtimeConfig = await response.json();
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

  return (
    <LocaleProvider locale="en" messageByID={MESSAGES}>
      <ApolloProvider client={client}>
        <div className={styles.root}>{children}</div>
      </ApolloProvider>
    </LocaleProvider>
  );
};

export default ReactApp;
