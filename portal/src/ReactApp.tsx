import React, { useEffect, useState } from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import authgear from "@authgear/web";
import Authenticated from "./Authenticated";
import ShowLoading from "./ShowLoading";
import AppsScreen from "./AppsScreen";
import AppScreen from "./AppScreen";

// ReactAppRoutes defines the routes.
const ReactAppRoutes: React.FC = function ReactAppRoutes() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/apps" replace={true} />} />
        <Route path="/apps" element={<AppsScreen />} />
        <Route path="/apps/:appID" element={<AppScreen />} />
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

  if (error != null) {
    return (
      <p>Failed to initialize the app. Please refresh this page to retry.</p>
    );
  }

  if (loading) {
    return <ShowLoading />;
  }

  return (
    <Authenticated>
      <ReactAppRoutes />
    </Authenticated>
  );
};

export default ReactApp;
