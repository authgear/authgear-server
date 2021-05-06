import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useNavigate } from "react-router-dom";
import { ICommandBarItemProps, SearchBox } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { gql, useLazyQuery, QueryLazyOptions } from "@apollo/client";
import NavBreadcrumb from "../../NavBreadcrumb";
import UsersList from "./UsersList";
import CommandBarContainer from "../../CommandBarContainer";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { encodeOffsetToCursor } from "../../util/pagination";
import {
  UsersListQuery,
  UsersListQueryVariables,
} from "./__generated__/UsersListQuery";
import {
  UsersSearchQuery,
  UsersSearchQueryVariables,
} from "./__generated__/UsersSearchQuery";
import { SearchUsersSortBy, SortDirection } from "./__generated__/globalTypes";
import ShowError from "../../ShowError";

import styles from "./UsersScreen.module.scss";

const pageSize = 10;

const LIST_QUERY = gql`
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

const SEARCH_QUERY = gql`
  query UsersSearchQuery(
    $searchKeyword: String!
    $pageSize: Int!
    $cursor: String
    $sortBy: SearchUsersSortBy
    $sortDirection: SortDirection
  ) {
    users: searchUsers(
      first: $pageSize
      after: $cursor
      searchKeyword: $searchKeyword
      sortBy: $sortBy
      sortDirection: $sortDirection
    ) {
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

const UsersScreen: React.FC = function UsersScreen() {
  const { searchEnabled } = useSystemConfig();
  const [searchKeyword, setSearchKeyword] = useState("");
  const [offset, setOffset] = useState(0);
  const [sortBy, setSortBy] = useState<SearchUsersSortBy | undefined>(
    undefined
  );
  const [sortDirection, setSortDirection] = useState<SortDirection | undefined>(
    undefined
  );

  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="UsersScreen.title" /> }];
  }, []);

  const onChangeSearchKeyword = useCallback((_e, value) => {
    if (value != null) {
      setSearchKeyword(value);
      // Reset offset when search keyword was changed.
      setOffset(0);
      // Reset sort when we are not searching.
      if (value === "") {
        setSortBy(undefined);
        setSortDirection(undefined);
      }
    }
  }, []);

  const commandBarItems: ICommandBarItemProps[] = useMemo(() => {
    if (searchEnabled) {
      return [
        {
          key: "search",
          onRender: () => {
            return (
              <SearchBox
                className={styles.searchBox}
                placeholder={renderToString("search")}
                value={searchKeyword}
                onChange={onChangeSearchKeyword}
              />
            );
          },
        },
      ];
    }
    return [];
  }, [renderToString, onChangeSearchKeyword, searchKeyword, searchEnabled]);

  const commandBarFarItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "addUser",
        text: renderToString("UsersScreen.add-user"),
        iconProps: { iconName: "CirclePlus" },
        onClick: () => navigate("./add-user"),
      },
    ];
  }, [navigate, renderToString]);

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

  const listQuery = useLazyQuery<UsersListQuery, UsersListQueryVariables>(
    LIST_QUERY,
    {
      fetchPolicy: "network-only",
    }
  );

  const searchQuery = useLazyQuery<UsersSearchQuery, UsersSearchQueryVariables>(
    SEARCH_QUERY,
    {
      fetchPolicy: "network-only",
    }
  );

  let execute: (options?: QueryLazyOptions<UsersSearchQueryVariables>) => void;
  let refetch: (() => void) | undefined;
  let loading: boolean;
  let error: unknown;
  let data: UsersListQuery | undefined;
  if (searchKeyword !== "") {
    execute = searchQuery[0];
    data = searchQuery[1].data;
    refetch = searchQuery[1].refetch;
    loading = searchQuery[1].loading;
    error = searchQuery[1].error;
  } else {
    execute = listQuery[0];
    data = listQuery[1].data;
    refetch = listQuery[1].refetch;
    loading = listQuery[1].loading;
    error = listQuery[1].error;
  }

  const prevDataRef = useRef<UsersListQuery | undefined>();
  useEffect(() => {
    prevDataRef.current = data;
  });
  const prevData = prevDataRef.current;

  const search = useCallback(() => {
    execute({
      variables: {
        sortBy,
        sortDirection,
        searchKeyword,
        pageSize,
        cursor,
      },
    });
  }, [execute, sortBy, sortDirection, searchKeyword, cursor]);

  // Initial execute
  useEffect(() => {
    search();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Execute when cursor, sortBy, or sortDirection changes.
  useEffect(() => {
    search();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [cursor, sortBy, sortDirection]);

  // Debounced execute when searchKeyword changes.
  const setTimeoutToken = useRef<ReturnType<typeof setTimeout> | undefined>();
  useEffect(() => {
    const token = setTimeout(() => {
      search();
    }, 500);

    setTimeoutToken.current = token;

    return () => {
      if (setTimeoutToken.current != null) {
        clearTimeout(setTimeoutToken.current);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchKeyword]);

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    return null;
  }, [error, refetch]);

  const onColumnClick = useCallback(
    (columnKey: SearchUsersSortBy) => {
      // Sort is not supported when we are not searching.
      if (searchKeyword === "") {
        setSortBy(undefined);
        setSortDirection(undefined);
        return;
      }

      if (sortBy === columnKey) {
        if (sortDirection == null) {
          setSortDirection(SortDirection.DESC);
        } else if (sortDirection === SortDirection.DESC) {
          setSortDirection(SortDirection.ASC);
        } else {
          setSortBy(undefined);
          setSortDirection(undefined);
        }
      } else {
        setSortBy(columnKey);
        setSortDirection(SortDirection.DESC);
      }
    },
    [searchKeyword, sortBy, sortDirection]
  );

  return (
    <CommandBarContainer
      isLoading={loading}
      className={styles.root}
      items={commandBarItems}
      farItems={commandBarFarItems}
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
          onColumnClick={onColumnClick}
          sortBy={searchKeyword === "" ? undefined : sortBy}
          sortDirection={searchKeyword === "" ? undefined : sortDirection}
        />
      </main>
    </CommandBarContainer>
  );
};

export default UsersScreen;
