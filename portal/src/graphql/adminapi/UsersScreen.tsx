import React, { useContext } from "react";
import { QueryRenderer } from "react-relay";
import { RelayUsersList, query } from "./UsersList";
import {
  UsersListQueryVariables,
  UsersListQueryResponse,
} from "./__generated__/UsersListQuery.graphql";
import AppContext from "../../AppContext";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";

const UsersScreen: React.FC = function UsersScreen() {
  const environment = useContext(AppContext);
  const pageSize = 1;
  return (
    <QueryRenderer<{
      variables: UsersListQueryVariables;
      response: UsersListQueryResponse;
    }>
      environment={environment}
      query={query}
      variables={{ pageSize }}
      render={({ error, props, retry }) => {
        if (error != null) {
          return <ShowError error={error} onRetry={retry} />;
        }
        if (props == null) {
          // FIXME(portal): Use Skimmer
          return <ShowLoading />;
        }
        return <RelayUsersList users={props} pageSize={1} />;
      }}
    />
  );
};

export default UsersScreen;
