import React from "react";
import { graphql, QueryRenderer } from "react-relay";
import { useParams } from "react-router-dom";
import { AppScreenQueryResponse } from "./__generated__/AppScreenQuery.graphql";
import { environment } from "./relay";
import ScreenHeader from "../../ScreenHeader";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";

const query = graphql`
  query AppScreenQuery($id: ID!) {
    node(id: $id) {
      ... on App {
        id
        appConfig
        secretConfig
      }
    }
  }
`;

interface Variables {
  id: string;
}

const ShowApp: React.FC<AppScreenQueryResponse> = function ShowApp(
  props: AppScreenQueryResponse
) {
  return <pre>{JSON.stringify(props.node, null, 2)}</pre>;
};

const AppScreen: React.FC = function AppScreen() {
  const { appID } = useParams();
  return (
    <div>
      <ScreenHeader />
      <QueryRenderer<{ variables: Variables; response: AppScreenQueryResponse }>
        environment={environment}
        query={query}
        variables={{ id: appID }}
        render={({ error, props }) => {
          if (error != null) {
            return <ShowError error={error} />;
          }
          if (props == null) {
            return <ShowLoading />;
          }
          return <ShowApp {...props} />;
        }}
      />
    </div>
  );
};

export default AppScreen;
