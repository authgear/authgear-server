import React, { useCallback, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Flex, Text } from "@radix-ui/themes";
import { PlusIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "../../intl";
import { useQuery } from "@apollo/client";
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
import { useDebounced } from "../../hook/useDebounced";

import styles from "./UsersScreen.module.css";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { Callout } from "../../components/v2/Callout/Callout";
import {
  UsersFilter,
  UsersFilterBar,
} from "../../components/users/UsersFilterBar";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "./query/rolesListQuery.generated";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "./query/groupsListQuery.generated";

const pageSize = 10;
// We have performance problem on the users query
// limit to 50 items for now
const searchResultSize = 50;

function useRemoteData(options: {
  filters: UsersFilter;
  offset: number;
  sortBy: UserSortBy | undefined;
  sortDirection: SortDirection | undefined;
}) {
  const { filters, offset, sortBy, sortDirection } = options;

  const isSearch =
    filters.searchKeyword !== "" ||
    filters.role != null ||
    filters.group != null;

  const [debouncedSearchKey] = useDebounced(filters.searchKeyword, 500);

  const filterGroupKeys = useMemo(() => {
    return filters.group == null ? undefined : [filters.group.group.key];
  }, [filters.group]);

  const filterRoleKeys = useMemo(() => {
    return filters.role == null ? undefined : [filters.role.role.key];
  }, [filters.role]);

  const cursor = useMemo(() => {
    if (isSearch) {
      // Search always query all rows.
      return undefined;
    }
    return encodeOffsetToCursor(offset);
  }, [isSearch, offset]);

  const {
    data,
    error,
    loading,
    refetch: refetchUsersListData,
  } = useQuery<UsersListQueryQuery, UsersListQueryQueryVariables>(
    UsersListQueryDocument,
    {
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
    }
  );

  const {
    data: rolesListData,
    loading: isRolesListDataLoading,
    error: rolesListDataError,
    refetch: refetchRolesListData,
  } = useQuery<RolesListQueryQuery, RolesListQueryQueryVariables>(
    RolesListQueryDocument,
    {
      variables: {
        pageSize: 0,
        searchKeyword: "",
      },
      fetchPolicy: "network-only",
    }
  );

  const {
    data: groupsListData,
    loading: isGroupsListDataLoading,
    error: groupsListDataError,
    refetch: refetchGroupsListData,
  } = useQuery<GroupsListQueryQuery, GroupsListQueryQueryVariables>(
    GroupsListQueryDocument,
    {
      variables: {
        pageSize: 0,
        searchKeyword: "",
      },
      fetchPolicy: "network-only",
    }
  );

  return {
    isLoading: loading || isRolesListDataLoading || isGroupsListDataLoading,
    isSearch: isSearch,
    error: error ?? rolesListDataError ?? groupsListDataError,
    data,
    isGroupsEmpty:
      groupsListData?.groups == null || groupsListData.groups.totalCount === 0,
    isRolesEmpty:
      rolesListData?.roles == null || rolesListData.roles.totalCount === 0,
    refetch: useCallback(
      async () =>
        Promise.all([
          refetchUsersListData(),
          refetchGroupsListData(),
          refetchRolesListData(),
        ]),
      [refetchGroupsListData, refetchRolesListData, refetchUsersListData]
    ),
  };
}

const UsersScreen: React.VFC = function UsersScreen() {
  const { searchEnabled } = useSystemConfig();

  const [filters, setFilters] = useState<UsersFilter>({
    searchKeyword: "",
    role: null,
    group: null,
  });

  const [offset, setOffset] = useState(0);
  const [sortBy, setSortBy] = useState<UserSortBy | undefined>(undefined);
  const [sortDirection, setSortDirection] = useState<SortDirection | undefined>(
    undefined
  );

  const navigate = useNavigate();

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const {
    data,
    isLoading,
    isGroupsEmpty,
    isRolesEmpty,
    error,
    refetch,
    isSearch,
  } = useRemoteData({ filters, offset, sortBy, sortDirection });

  // Count the length instead of totalCount, because totalCount is not available for pgsearch
  const isTotalExceededLimit =
    (data?.users?.edges?.length ?? 0) >= searchResultSize;

  const messageBar = useMemo(() => {
    if (error != null) {
      return (
        <ShowError
          error={error}
          onRetry={() => {
            refetch().finally(() => {});
          }}
        />
      );
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
      isLoading={isLoading}
      messageBar={messageBar}
      hideCommandBar={true}
    >
      <ScreenContent className={styles.content} layout="list">
        <div className={styles.widget}>
          <div className={styles.pageTitleRow}>
            <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
              <FormattedMessage id="UsersScreen.title" />
            </Text>
            <PrimaryButton
              size="2"
              text={
                <Flex align="center" gap="2">
                  <PlusIcon width="1rem" height="1rem" />
                  <FormattedMessage id="UsersScreen.add-user" />
                </Flex>
              }
              onClick={useCallback(() => {
                navigate("./add-user");
              }, [navigate])}
            />
          </div>
          <UsersFilterBar
            className={styles.filterBar}
            showSearchBar={searchEnabled}
            showGroupFilter={!isGroupsEmpty}
            showRoleFilter={!isRolesEmpty}
            filters={filters}
            onFilterChange={setFilters}
          />
          {isSearch && isTotalExceededLimit && !isLoading ? (
            <Callout
              className={styles.message}
              type="info"
              showCloseButton={false}
              text={<FormattedMessage id="UsersScreen.search.resultLimited" />}
            />
          ) : null}
        </div>
        <UsersList
          className={styles.widget}
          isSearch={isSearch}
          loading={isLoading}
          users={data?.users ?? null}
          offset={offset}
          pageSize={pageSize}
          totalCount={data?.users?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
          onColumnClick={onColumnClick}
          sortBy={sortBy}
          sortDirection={sortDirection}
          showRolesAndGroups={!isRolesEmpty || !isGroupsEmpty}
        />
      </ScreenContent>
    </CommandBarContainer>
  );
};

export default UsersScreen;
