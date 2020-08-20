import React from "react";
import { graphql, QueryRenderer } from "react-relay";
import { environment } from "./relay";
import { AppQueryResponse } from "./__generated__/AppQuery.graphql";

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
  return <div>{props.viewer?.id}</div>;
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
