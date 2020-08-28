import React, { useContext } from "react";
import { graphql, QueryRenderer } from "react-relay";
import { UsersScreenQueryResponse } from "./__generated__/UsersScreenQuery.graphql";
import AppContext from "../../AppContext";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";

const query = graphql`
  query UsersScreenQuery {
    users {
      edges {
        node {
          id
          createdAt
        }
      }
    }
  }
`;

interface Empty {}

const UsersScreen: React.FC = function UsersScreen() {
  // FIXME(portal): Use Pagination Container
  // FIXME(portal): Use DetailsList
  const environment = useContext(AppContext);
  return (
    <QueryRenderer<{ variables: Empty; response: UsersScreenQueryResponse }>
      environment={environment}
      query={query}
      variables={{}}
      render={({ error, props, retry }) => {
        if (error != null) {
          return <ShowError error={error} onRetry={retry} />;
        }
        if (props == null) {
          return <ShowLoading />;
        }
        return null;
      }}
    />
  );
};

export default UsersScreen;
