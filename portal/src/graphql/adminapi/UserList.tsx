import React, { useContext } from "react";
import { graphql, QueryRenderer } from "react-relay";
import { UserListQueryResponse } from "./__generated__/UserListQuery.graphql";
import AppContext from "../../AppContext";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";

const query = graphql`
  query UserListQuery {
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

const RawUserList: React.FC<UserListQueryResponse> = function RawUserList(
  props: UserListQueryResponse
) {
  return <pre>{JSON.stringify(props.users, null, 2)}</pre>;
};

interface Empty {}

const UserList: React.FC = function UserList() {
  const environment = useContext(AppContext);
  return (
    <QueryRenderer<{ variables: Empty; response: UserListQueryResponse }>
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
        return <RawUserList {...props} />;
      }}
    />
  );
};

export default UserList;
