import React, { useEffect, useState, useCallback } from "react";
import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import authgear from "@authgear/web";
import Authenticated from "./Authenticated";

const ShowLoading: React.FC = function ShowLoading() {
  return <div>Loading...</div>;
};

const Root: React.FC = function Root() {
  const redirectURI = window.location.origin + "/";

  const onClickLogout = useCallback(() => {
    authgear
      .logout({
        redirectURI,
      })
      .catch((err) => {
        console.error(err);
      });
  }, [redirectURI]);

  return (
    <div>
      <p>This is /</p>
      <Link to="/apps">Go to /apps</Link>
      <button type="button" onClick={onClickLogout}>
        Click here to logout
      </button>
    </div>
  );
};

const Apps: React.FC = function Apps() {
  const redirectURI = window.location.origin + "/";

  const onClickLogout = useCallback(() => {
    authgear
      .logout({
        redirectURI,
      })
      .catch((err) => {
        console.error(err);
      });
  }, [redirectURI]);

  return (
    <div>
      <p>This is /apps</p>
      <Link to="/">Go to /</Link>
      <button type="button" onClick={onClickLogout}>
        Click here to logout
      </button>
    </div>
  );
};

const App: React.FC = function App() {
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
    <BrowserRouter>
      <Routes>
        <Authenticated>
          <Route path="/" element={<Root />} />
        </Authenticated>
        <Authenticated>
          <Route path="/apps" element={<Apps />} />
        </Authenticated>
      </Routes>
    </BrowserRouter>
  );
};

export default App;
