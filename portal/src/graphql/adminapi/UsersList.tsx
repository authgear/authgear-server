import React, {
  useMemo,
  useContext,
  useState,
  useCallback,
  useRef,
  useEffect,
} from "react";
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
  UsersListQuery_users_edges_node_identities,
} from "./__generated__/UsersListQuery";

import ShowError from "../../ShowError";
import PaginationWidget from "../../PaginationWidget";

import { encodeOffsetToCursor } from "../../util/pagination";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UsersList.module.scss";

interface Props {
  loading: boolean;
  users: UsersListQuery_users | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface UserInfo {
  username: string;
  phone: string;
  email: string;
}

interface IdentityClaims extends GQL_JSONObject {
  email?: string;
  preferred_username?: string;
  phone_number?: string;
}

function extractUserInfoFromIdentities(
  identities: UsersListQuery_users_edges_node_identities | null
): UserInfo {
  const placeholder = "-";
  const claimsList: IdentityClaims[] = [];

  (identities?.edges ?? []).forEach((edge) => {
    if (edge?.node?.claims == null) {
      return;
    }
    claimsList.push(edge.node.claims);
  });

  // TODO: consider update UI design to allow multiple email, username and phone number
  const email =
    claimsList.map((claims) => claims.email).filter(Boolean)[0] ?? placeholder;
  const username =
    claimsList.map((claims) => claims.preferred_username).filter(Boolean)[0] ??
    placeholder;
  const phone =
    claimsList.map((claims) => claims.phone_number).filter(Boolean)[0] ??
    placeholder;

  return { email, username, phone };
}

const PlainUsersList: React.FC<Props> = function PlainUsersList(props: Props) {
  const { loading, offset, pageSize, totalCount, onChangeOffset } = props;
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
      key: "signedUp",
      fieldName: "signedUp",
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

  const items: {
    id: string;
    signedUp: string | null;
    username: string;
    phone: string;
    email: string;
    lastLoginAt: string | null;
  }[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          const userInfo = extractUserInfoFromIdentities(node.identities);
          items.push({
            id: node.id,
            signedUp: formatDatetime(locale, node.createdAt),
            lastLoginAt: formatDatetime(locale, node.lastLoginAt),
            ...userInfo,
          });
        }
      }
    }
    return items;
  }, [edges, locale]);

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
