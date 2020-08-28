import React, { useMemo, useContext } from "react";
import { graphql, RelayRefetchProp, createRefetchContainer } from "react-relay";
import {
  DetailsList,
  DetailsListLayoutMode,
  SelectionMode,
  IColumn,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { UsersList_users } from "./__generated__/UsersList_users.graphql";
import styles from "./UsersList.module.scss";

interface Props {
  users: UsersList_users;
  pageSize: number;
  relay: RelayRefetchProp;
}

const UsersList: React.FC<Props> = function UsersList(props: Props) {
  const { renderToString } = useContext(Context);

  const columns: IColumn[] = [
    {
      key: "id",
      fieldName: "id",
      name: renderToString("UsersList.column.id"),
      minWidth: 400,
      maxWidth: 400,
    },
    {
      key: "createdAt",
      fieldName: "createdAt",
      name: renderToString("UsersList.column.created-at"),
      minWidth: 300,
    },
  ];

  const edges = props.users.users?.edges;

  const items: {
    id: string;
    createdAt: unknown;
  }[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          items.push({
            id: node.id,
            createdAt: node.createdAt,
          });
        }
      }
    }
    return items;
  }, [edges]);

  return (
    <div className={styles.root}>
      <DetailsList
        selectionMode={SelectionMode.none}
        layoutMode={DetailsListLayoutMode.justified}
        columns={columns}
        items={items}
      />
    </div>
  );
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
