import React, {
  useMemo,
  useContext,
  useState,
  useCallback,
  useRef,
  useEffect,
} from "react";
import cn from "classnames";
import { useQuery, gql } from "@apollo/client";
import {
  ShimmeredDetailsList,
  DetailsListLayoutMode,
  SelectionMode,
  IColumn,
  IDetailsRowProps,
  DetailsRow,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { Link } from "react-router-dom";
import {
  UsersListQuery,
  UsersListQuery_users,
  UsersListQueryVariables,
} from "./__generated__/UsersListQuery";

import ShowError from "../../ShowError";
import PaginationWidget from "../../PaginationWidget";

import { encodeOffsetToCursor } from "../../util/pagination";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UsersList.module.scss";
import { extractUserInfoFromIdentities } from "../../util/user";
import { nonNullable } from "../../util/types";

interface PlainUsersListProps {
  className?: string;
  loading: boolean;
  users: UsersListQuery_users | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface UserListItem {
  id: string;
  createdAt: string | null;
  username: string;
  phone: string;
  email: string;
  lastLoginAt: string | null;
}

const isUserListItem = (value: any): value is UserListItem => {
  if (!(value instanceof Object)) {
    return false;
  }
  return (
    "id" in value && "username" in value && "phone" in value && "email" in value
  );
};

const PlainUsersList: React.FC<PlainUsersListProps> = function PlainUsersList(
  props: PlainUsersListProps
) {
  const {
    className,
    loading,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  } = props;
  const edges = props.users?.edges;

  const { renderToString, locale } = useContext(Context);

  const columns: IColumn[] = [
    {
      key: "id",
      fieldName: "id",
      name: renderToString("UsersList.column.id"),
      minWidth: 400,
      maxWidth: 400,
    },
    {
      key: "username",
      fieldName: "username",
      name: renderToString("UsersList.column.username"),
      minWidth: 150,
    },
    {
      key: "email",
      fieldName: "email",
      name: renderToString("UsersList.column.email"),
      minWidth: 150,
    },
    {
      key: "phone",
      fieldName: "phone",
      name: renderToString("UsersList.column.phone"),
      minWidth: 150,
    },
    {
      key: "createdAt",
      fieldName: "createdAt",
      name: renderToString("UsersList.column.signed-up"),
      minWidth: 200,
    },
    {
      key: "lastLoginAt",
      fieldName: "lastLoginAt",
      name: renderToString("UsersList.column.last-login-at"),
      minWidth: 200,
    },
  ];

  // TODO: consider update UI design to allow multiple email, username and phone number
  const items: UserListItem[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          const identities =
            node.identities?.edges
              ?.map((edge) => edge?.node)
              ?.filter(nonNullable) ?? [];
          const userInfo = extractUserInfoFromIdentities(identities);
          const placeholder = "-";
          items.push({
            id: node.id,
            createdAt: formatDatetime(locale, node.createdAt),
            lastLoginAt: formatDatetime(locale, node.lastLoginAt),
            username: userInfo.username ?? placeholder,
            phone: userInfo.phone ?? placeholder,
            email: userInfo.email ?? placeholder,
          });
        }
      }
    }
    return items;
  }, [edges, locale]);

  const onRenderUserRow = React.useCallback((props?: IDetailsRowProps) => {
    if (props == null) {
      return null;
    }
    const targetPath = isUserListItem(props.item)
      ? `./${props.item.id}/details`
      : ".";
    return (
      <Link to={targetPath}>
        <DetailsRow {...props} />
      </Link>
    );
  }, []);

  return (
    <div className={cn(styles.root, className)}>
      <ShimmeredDetailsList
        enableShimmer={loading}
        onRenderRow={onRenderUserRow}
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
          lastLoginAt
          identities {
            edges {
              node {
                id
                claims
              }
            }
          }
        }
      }
      totalCount
    }
  }
`;

const pageSize = 10;

interface Props {
  className?: string;
}

const UsersList: React.FC<Props> = function UsersList(props: Props) {
  const { className } = props;
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

  const prevDataRef = useRef<UsersListQuery | undefined>();
  useEffect(() => {
    prevDataRef.current = data;
  });
  const prevData = prevDataRef.current;

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <PlainUsersList
      className={className}
      loading={loading}
      users={data?.users ?? null}
      offset={offset}
      pageSize={pageSize}
      totalCount={(data ?? prevData)?.users?.totalCount ?? undefined}
      onChangeOffset={onChangeOffset}
    />
  );
};

export default UsersList;
