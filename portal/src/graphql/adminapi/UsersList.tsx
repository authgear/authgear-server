import React from "react";
import { graphql, RelayRefetchProp, createRefetchContainer } from "react-relay";
import { UsersList_users } from "./__generated__/UsersList_users.graphql";

interface Props {
  users: UsersList_users;
  pageSize: number;
  relay: RelayRefetchProp;
}

const UsersList: React.FC<Props> = function UsersList(props: Props) {
  // FIXME(portal): Use DetailsList
  console.log("louis#props", props);
  return null;
};

const query = graphql`
  query UsersListQuery($pageSize: Int!, $cursor: String) {
    ...UsersList_users
  }
`;

// There is createRefetchContainer and createPaginationContainer.
// But the hasMore() createPaginationContainer for some reason always return false even hasNextPage is true.
// So we use createRefetchContainer instead.
const RelayUsersList = createRefetchContainer<Exclude<Props, "relay">>(
  UsersList,
  {
    users: graphql`
      fragment UsersList_users on Query {
        users(first: $pageSize, after: $cursor)
          @connection(key: "UsersList_users") {
          edges {
            node {
              id
              createdAt
            }
          }
          pageInfo {
            hasNextPage
            endCursor
          }
          totalCount
        }
      }
    `,
  },
  query
);

export { RelayUsersList, query };
