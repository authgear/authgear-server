import React, { useMemo, useContext, useState, useCallback } from "react";
import { useQuery, gql } from "@apollo/client";
import {
  ShimmeredDetailsList,
  DetailsListLayoutMode,
  SelectionMode,
  IColumn,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import {
  UsersListQuery,
  UsersListQuery_users,
  UsersListQueryVariables,
} from "./__generated__/UsersListQuery";
import ShowError from "../../ShowError";
import PaginationWidget from "../../PaginationWidget";
import { encodeOffsetToCursor } from "../../util/pagination";
import styles from "./UsersList.module.scss";

interface Props {
  loading: boolean;
  users: UsersListQuery_users | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

const PlainUsersList: React.FC<Props> = function PlainUsersList(props: Props) {
  const { loading, offset, pageSize, totalCount, onChangeOffset } = props;
  const edges = props.users?.edges;

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
      <ShimmeredDetailsList
        enableShimmer={loading}
        selectionMode={SelectionMode.none}
        layoutMode={DetailsListLayoutMode.justified}
        columns={columns}
        items={items}
      />
      <PaginationWidget
        className={styles.pagination}
        offset={offset}
        pageSize={pageSize}
        totalCount={totalCount}
        onChangeOffset={onChangeOffset}
      />
    </div>
  );
};

const query = gql`
  query UsersListQuery($pageSize: Int!, $cursor: String) {
    users(first: $pageSize, after: $cursor) {
      edges {
        node {
          id
          createdAt
        }
      }
      totalCount
    }
  }
`;

const pageSize = 10;

const UsersList: React.FC = function UsersList() {
  const [offset, setOffset] = useState(0);

  // after: is exclusive so if we pass it "offset:0",
  // The first item is excluded.
  // Therefore we have adjust it by -1.
  const cursor = useMemo(() => {
    if (offset === 0) {
      return null;
    }
    return encodeOffsetToCursor(offset - 1);
  }, [offset]);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const { loading, error, data, refetch } = useQuery<
    UsersListQuery,
    UsersListQueryVariables
  >(query, {
    variables: {
      pageSize,
      cursor,
    },
  });

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <PlainUsersList
      loading={loading}
      users={data?.users ?? null}
      offset={offset}
      pageSize={pageSize}
      totalCount={data?.users?.totalCount ?? undefined}
      onChangeOffset={onChangeOffset}
    />
  );
};

export default UsersList;
