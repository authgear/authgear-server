import React, { useCallback, useEffect, useState } from "react";
import { graphql, QueryRenderer } from "react-relay";
import authgear from "@authgear/web";
import { environment } from "./relay";
import { AppQueryResponse } from "./__generated__/AppQuery.graphql";
import styles from "./App.module.scss";

const query = graphql`
  query AppQuery {
    viewer {
      id
    }
  }
`;

const ShowQueryResult: React.FC<AppQueryResponse> = function ShowQueryResult(
  props: AppQueryResponse
) {
  const { viewer } = props;
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

  useEffect(() => {
    if (viewer == null) {
      // Normally we should call endAuthorization after being redirected back to here.
      // But we know that we are first party app and are using response_type=none so
      // we can skip that.
      authgear
        .startAuthorization({
          redirectURI,
          prompt: "login",
        })
        .catch((err) => {
          console.error(err);
        });
    }
  }, [viewer, redirectURI]);

  if (viewer != null) {
    return (
      <div className={styles.app}>
        <p>You are logged in as {viewer.id}</p>
        <button
          type="button"
          className={styles.logoutButton}
          onClick={onClickLogout}
        >
          Click here to logout
        </button>
      </div>
    );
  }

  return null;
};

interface ShowErrorProps {
  error: unknown;
}

const ShowError: React.FC<ShowErrorProps> = function ShowError(
  props: ShowErrorProps
) {
  const { error } = props;
  if (error instanceof Error) {
    return (
      <div
        style={{
          whiteSpace: "pre",
        }}
      >
        {error.name}: {error.message}
        <br /> {error.stack}
      </div>
    );
  }
  return <div>Non-Error error: {String(error)}</div>;
};

const ShowLoading: React.FC = function ShowLoading() {
  return <div>Loading...</div>;
};

interface Empty {}

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
    <QueryRenderer<{ variables: Empty; response: AppQueryResponse }>
      environment={environment}
      query={query}
      variables={{}}
      render={({ error, props }) => {
        if (error != null) {
          return <ShowError error={error} />;
        }
        if (props == null) {
          return <ShowLoading />;
        }
        return <ShowQueryResult {...props} />;
      }}
    />
  );
};

export default App;
