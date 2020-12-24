import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useNavigate } from "react-router-dom";
import { ICommandBarItemProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { gql, useQuery } from "@apollo/client";
import NavBreadcrumb from "../../NavBreadcrumb";
import UsersList from "./UsersList";
import CommandBarContainer from "../../CommandBarContainer";
import { encodeOffsetToCursor } from "../../util/pagination";
import {
  UsersListQuery,
  UsersListQueryVariables,
} from "./__generated__/UsersListQuery";
import ShowError from "../../ShowError";

import styles from "./UsersScreen.module.scss";

const query = gql`
  query UsersListQuery($pageSize: Int!, $cursor: String) {
    users(first: $pageSize, after: $cursor) {
      edges {
        node {
          id
          createdAt
          lastLoginAt
          isDisabled
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

const UsersScreen: React.FC = function UsersScreen() {
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="UsersScreen.title" /> }];
  }, []);

  const commandBarItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "addUser",
        text: renderToString("UsersScreen.add-user"),
        iconProps: { iconName: "CirclePlus" },
        onClick: () => navigate("./add-user"),
      },
    ];
  }, [navigate, renderToString]);

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
    fetchPolicy: "network-only",
  });

  const prevDataRef = useRef<UsersListQuery | undefined>();
  useEffect(() => {
    prevDataRef.current = data;
  });
  const prevData = prevDataRef.current;

  const messageBar = useMemo(
    () => error && <ShowError error={error} onRetry={refetch} />,
    [error, refetch]
  );

  return (
    <CommandBarContainer
      isLoading={loading}
      className={styles.root}
      farItems={commandBarItems}
      messageBar={messageBar}
    >
      <main className={styles.content}>
        <NavBreadcrumb items={items} />
        <UsersList
          className={styles.usersList}
          loading={loading}
          users={data?.users ?? null}
          offset={offset}
          pageSize={pageSize}
          totalCount={(data ?? prevData)?.users?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
        />
      </main>
    </CommandBarContainer>
  );
};

export default UsersScreen;
