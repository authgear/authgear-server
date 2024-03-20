import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  ICommandBarItemProps,
  SearchBox,
  ISearchBoxProps,
  MessageBar,
} from "@fluentui/react";
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
import { onRenderCommandBarPrimaryButton } from "../../CommandBarPrimaryButton";
import {
  GroupsFilterDropdown,
  GroupsFilterDropdownOption,
} from "../../components/users/GroupsFilterDropdown";

const pageSize = 10;
// We have performance problem on the users query
// limit to 10 items for now
const searchResultSize = 10;

function useOnClearFilterCallback<T>(setT: (value: T | null) => void) {
  return useCallback(() => setT(null), [setT]);
}

const UsersScreen: React.VFC = function UsersScreen() {
  const { searchEnabled } = useSystemConfig();

  const [searchKeyword, setSearchKeyword] = useState("");
  const [groupFilter, setGroupFilter] =
    useState<GroupsFilterDropdownOption | null>(null);
  const clearGroupFilter = useOnClearFilterCallback(setGroupFilter);
  const debouncedSearchKey = useDelayedValue(searchKeyword, 500);

  const [offset, setOffset] = useState(0);
  const [sortBy, setSortBy] = useState<UserSortBy | undefined>(undefined);
  const [sortDirection, setSortDirection] = useState<SortDirection | undefined>(
    undefined
  );

  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  const isSearch = searchKeyword !== "";

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

  const searchBoxProps: ISearchBoxProps = useMemo(() => {
    return {
      className: styles.searchBox,
      placeholder: renderToString("search"),
      value: searchKeyword,
      onChange: onChangeSearchKeyword,
      onClear: onClearSearchKeyword,
    };
  }, [
    renderToString,
    searchKeyword,
    onChangeSearchKeyword,
    onClearSearchKeyword,
  ]);

  // If secondaryItems changes on every key stroke,
  // input method such as cangjie cannot be used.
  // Every key stroke is entered into the text box literally without giving us a chance to select character.
  // This can be work around by using context.
  const secondaryItems: ICommandBarItemProps[] = useMemo(() => {
    const items: ICommandBarItemProps[] = [];
    if (searchEnabled) {
      items.push({
        key: "search",
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: () => <SearchBox {...searchBoxProps} />,
      });
    }

    items.push({
      key: "groups-filter",
      // eslint-disable-next-line react/no-unstable-nested-components
      onRender: () => (
        <GroupsFilterDropdown
          value={groupFilter}
          onChange={setGroupFilter}
          onClear={clearGroupFilter}
        />
      ),
    });

    return items;
  }, [clearGroupFilter, groupFilter, searchBoxProps, searchEnabled]);

  const primaryItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "addUser",
        text: renderToString("UsersScreen.add-user"),
        iconProps: { iconName: "CirclePlus" },
        onClick: () => navigate("./add-user"),
        onRender: onRenderCommandBarPrimaryButton,
      },
    ];
  }, [navigate, renderToString]);

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
    return groupFilter == null ? undefined : [groupFilter.group.key];
  }, [groupFilter]);

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
      primaryItems={primaryItems}
      secondaryItems={secondaryItems}
      messageBar={messageBar}
    >
      <ScreenContent className={styles.content} layout="list">
        <div className={styles.widget}>
          <NavBreadcrumb className="block" items={items} />
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
