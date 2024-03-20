import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { MessageBar } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useQuery } from "@apollo/client";
import NavBreadcrumb from "../../NavBreadcrumb";
import UsersList from "./UsersList";
import CommandBarContainer from "../../CommandBarContainer";
import ScreenContent from "../../ScreenContent";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { encodeOffsetToCursor } from "../../util/pagination";
import {
  UsersListQueryQuery,
  UsersListQueryQueryVariables,
  UsersListQueryDocument,
} from "./query/usersListQuery.generated";
import { UserSortBy, SortDirection } from "./globalTypes.generated";
import ShowError from "../../ShowError";
import useDelayedValue from "../../hook/useDelayedValue";

import styles from "./UsersScreen.module.css";
import PrimaryButton from "../../PrimaryButton";
import {
  UsersFilter,
  UsersFilterBar,
} from "../../components/users/UsersFilterBar";

const pageSize = 10;
// We have performance problem on the users query
// limit to 10 items for now
const searchResultSize = 10;

const UsersScreen: React.VFC = function UsersScreen() {
  const { searchEnabled } = useSystemConfig();

  const [filters, setFilters] = useState<UsersFilter>({
    searchKeyword: "",
    role: null,
    group: null,
  });
  const debouncedSearchKey = useDelayedValue(filters.searchKeyword, 500);

  const [offset, setOffset] = useState(0);
  const [sortBy, setSortBy] = useState<UserSortBy | undefined>(undefined);
  const [sortDirection, setSortDirection] = useState<SortDirection | undefined>(
    undefined
  );

  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const isSearch = filters.searchKeyword !== "";

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="UsersScreen.title" /> }];
  }, []);

  // after: is exclusive so if we pass it "offset:0",
  // The first item is excluded.
  // Therefore we have adjust it by -1.
  const cursor = useMemo(() => {
    if (isSearch) {
      // Search always query all rows.
      return null;
    }
    if (offset === 0) {
      return null;
    }
    return encodeOffsetToCursor(offset - 1);
  }, [isSearch, offset]);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const filterGroupKeys = useMemo(() => {
    return filters.group == null ? undefined : [filters.group.group.key];
  }, [filters.group]);

  const filterRoleKeys = useMemo(() => {
    return filters.role == null ? undefined : [filters.role.role.key];
  }, [filters.role]);

  const { data, error, loading, refetch } = useQuery<
    UsersListQueryQuery,
    UsersListQueryQueryVariables
  >(UsersListQueryDocument, {
    variables: {
      pageSize: isSearch ? searchResultSize : pageSize,
      cursor,
      sortBy,
      sortDirection,
      searchKeyword: debouncedSearchKey,
      groupKeys: filterGroupKeys,
      roleKeys: filterRoleKeys,
    },
    fetchPolicy: "network-only",
  });

  const isTotalExceededLimit =
    (data?.users?.totalCount ?? 0) > searchResultSize;

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
          setSortDirection(SortDirection.Desc);
        } else if (sortDirection === SortDirection.Desc) {
          setSortDirection(SortDirection.Asc);
        } else {
          setSortBy(undefined);
          setSortDirection(undefined);
        }
      } else {
        setSortBy(columnKey);
        setSortDirection(SortDirection.Desc);
      }
    },
    [sortBy, sortDirection]
  );

  return (
    <CommandBarContainer
      className={styles.root}
      isLoading={loading}
      messageBar={messageBar}
      hideCommandBar={true}
    >
      <ScreenContent className={styles.content} layout="list">
        <div className={styles.widget}>
          <div className="flex gap-x-1">
            <NavBreadcrumb
              className="flex-1 overflow-hidden items-center"
              items={items}
            />
            <PrimaryButton
              text={renderToString("UsersScreen.add-user")}
              iconProps={useMemo(() => ({ iconName: "Add" }), [])}
              onClick={useCallback(() => navigate("./add-user"), [navigate])}
            />
          </div>
          <UsersFilterBar
            className="mt-12"
            showSearchBar={searchEnabled}
            filters={filters}
            onFilterChange={setFilters}
          />
          {isSearch && isTotalExceededLimit && !loading ? (
            <MessageBar className={styles.message}>
              <FormattedMessage id="UsersScreen.search.resultLimited" />
            </MessageBar>
          ) : null}
        </div>
        <UsersList
          className={styles.widget}
          isSearch={isSearch}
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
      </ScreenContent>
    </CommandBarContainer>
  );
};

export default UsersScreen;
