import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ICommandBarItemProps, SearchBox } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { gql, useQuery } from "@apollo/client";
import NavBreadcrumb from "../../NavBreadcrumb";
import UsersList from "./UsersList";
import CommandBarContainer from "../../CommandBarContainer";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { encodeOffsetToCursor } from "../../util/pagination";
import {
  UsersListQuery,
  UsersListQueryVariables,
} from "./__generated__/UsersListQuery";
import { UserSortBy, SortDirection } from "./__generated__/globalTypes";
import ShowError from "../../ShowError";
import useDelayedValue from "../../hook/useDelayedValue";

import styles from "./UsersScreen.module.scss";

const pageSize = 10;

const LIST_QUERY = gql`
  query UsersListQuery(
    $searchKeyword: String!
    $pageSize: Int!
    $cursor: String
    $sortBy: UserSortBy
    $sortDirection: SortDirection
  ) {
    users(
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
          standardAttributes
        }
      }
      totalCount
    }
  }
`;

const UsersScreen: React.FC = function UsersScreen() {
  const { searchEnabled } = useSystemConfig();

  const [searchKeyword, setSearchKeyword] = useState("");
  const debouncedSearchKey = useDelayedValue(searchKeyword, 500);

  const [offset, setOffset] = useState(0);
  const [sortBy, setSortBy] = useState<UserSortBy | undefined>(undefined);
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
    }
  }, []);

  const onClearSearchKeyword = useCallback((_e) => {
    setSearchKeyword("");
    // Reset offset when search keyword was changed.
    setOffset(0);
  }, []);

  const commandBarItems: ICommandBarItemProps[] = useMemo(() => {
    if (searchEnabled) {
      return [
        {
          key: "search",
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: () => {
            return (
              <SearchBox
                className={styles.searchBox}
                placeholder={renderToString("search")}
                value={searchKeyword}
                onChange={onChangeSearchKeyword}
                onClear={onClearSearchKeyword}
              />
            );
          },
        },
      ];
    }
    return [];
  }, [
    renderToString,
    onChangeSearchKeyword,
    searchKeyword,
    searchEnabled,
    onClearSearchKeyword,
  ]);

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

  const { data, error, loading, refetch } = useQuery<
    UsersListQuery,
    UsersListQueryVariables
  >(LIST_QUERY, {
    variables: {
      pageSize,
      cursor,
      sortBy,
      sortDirection,
      searchKeyword: debouncedSearchKey,
    },
    fetchPolicy: "network-only",
  });

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    return null;
  }, [error, refetch]);

  const onColumnClick = useCallback(
    (columnKey: UserSortBy) => {
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
    [sortBy, sortDirection]
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
          totalCount={data?.users?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
          onColumnClick={onColumnClick}
          sortBy={sortBy}
          sortDirection={sortDirection}
        />
      </main>
    </CommandBarContainer>
  );
};

export default UsersScreen;
