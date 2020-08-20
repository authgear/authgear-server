import React, { useCallback, useEffect } from "react";
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
  const onClickLogout = useCallback(() => {
    authgear
      .logout({
        redirectURI: window.location.href,
      })
      .catch((err) => {
        console.error(err);
      });
  }, []);
  useEffect(() => {
    if (viewer == null) {
      // Normally we should call endAuthorization after being redirected back to here.
      // But we know that we are first party app and are using response_type=none so
      // we can skip that.
      authgear
        .startAuthorization({
          prompt: "login",
          redirectURI: window.location.href,
        })
        .catch((err) => {
          console.error(err);
        });
    }
  }, [viewer]);

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
  return (
    <QueryRenderer<{ variables: Empty; response: AppQueryResponse }>
      environment={environment}
      query={query}
      variables={{}}
      render={({ error, props }) => {
        if (error) {
          return <ShowError error={error} />;
        }
        if (!props) {
          return <ShowLoading />;
        }
        return <ShowQueryResult {...props} />;
      }}
    />
  );
};

export default App;
