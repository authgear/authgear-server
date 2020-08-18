import React from "react";
import { graphql, QueryRenderer } from "react-relay";
import { environment } from "./relay";
import { AppQueryResponse } from "./__generated__/AppQuery.graphql";

const query = graphql`
  query AppQuery {
    users {
      edges {
        node {
          id
        }
      }
    }
  }
`;

const ShowQueryResult: React.FC<AppQueryResponse> = function ShowQueryResult(
  props: AppQueryResponse
) {
  return (
    <div>
      {props.users.edges?.map((edge, i) => {
        return <div key={i}>{edge?.node?.id}</div>;
      })}
    </div>
  );
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

const App: React.FC = function App() {
  return (
    <QueryRenderer
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
        // @ts-expect-error
        return <ShowQueryResult {...props} />;
      }}
    />
  );
};

export default App;
