import { useQuery } from "@apollo/client";
import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { ISearchBoxProps, SearchBox, MessageBar } from "@fluentui/react";
import {
  GroupsListQueryDocument,
  GroupsListQueryQuery,
  GroupsListQueryQueryVariables,
} from "./query/groupsListQuery.generated";
import styles from "./GroupsScreen.module.css";
import { encodeOffsetToCursor } from "../../util/pagination";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import useDelayedValue from "../../hook/useDelayedValue";
import { RoleAndGroupsLayout } from "../../RoleAndGroupsLayout";
import { GroupsEmptyView } from "../../components/roles-and-groups/GroupsEmptyView";
import { ReactRouterLinkComponent } from "../../ReactRouterLink";
import { RolesAndGroupsEmptyView } from "../../components/roles-and-groups/RolesAndGroupsEmptyView";

const pageSize = 10;
const searchResultSize = -1;

const GroupsScreen: React.VFC = function GroupsScreen() {
  const { renderToString } = useContext(Context);
  const [searchKeyword, setSearchKeyword] = useState("");
  const { appID } = useParams<{ appID: string }>();

  const isSearch = searchKeyword !== "";
  const debouncedSearchKey = useDelayedValue(searchKeyword, 500);

  const [offset, setOffset] = useState(0);
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

  const { data, loading, previousData } = useQuery<
    GroupsListQueryQuery,
    GroupsListQueryQueryVariables
  >(GroupsListQueryDocument, {
    variables: {
      pageSize: isSearch ? searchResultSize : pageSize,
      searchKeyword: debouncedSearchKey,
      cursor,
    },
    fetchPolicy: "network-only",
  });

  const isInitialLoading = loading && previousData == null;

  const isEmpty = !isInitialLoading && data?.groups?.totalCount === 0;
  const isSearchEmpty = isSearch && data?.groups?.edges?.length === 0;

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

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="GroupsScreen.title" /> }];
  }, []);

  const headerSubItem = useMemo(() => {
    return !isEmpty ? (
      <ReactRouterLinkComponent
        component={RolesAndGroupsEmptyView.CreateButton}
        to={`/project/${appID}/user-management/groups/add-group`}
        text={<FormattedMessage id="GroupsEmptyView.button.text" />}
      />
    ) : null;
  }, [appID, isEmpty]);

  return (
    <RoleAndGroupsLayout
      headerBreadcrumbs={items}
      headerSubitem={headerSubItem}
    >
      {!isEmpty ? <SearchBox {...searchBoxProps} /> : null}
      {isEmpty ? (
        <GroupsEmptyView className={styles.emptyStateContainer} />
      ) : isSearchEmpty ? (
        <MessageBar className={cn(styles.message)}>
          <FormattedMessage id="GroupsScreen.empty.search" />
        </MessageBar>
      ) : (
        // TODO: GroupsList
        <GroupsEmptyView className={styles.emptyStateContainer} />
      )}
    </RoleAndGroupsLayout>
  );
};

export default GroupsScreen;
