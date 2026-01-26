import { useQuery } from "@apollo/client";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { ISearchBoxProps, SearchBox } from "@fluentui/react";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "./query/rolesListQuery.generated";
import styles from "./RolesScreen.module.css";
import { encodeOffsetToCursor } from "../../util/pagination";
import { Context, FormattedMessage } from "../../intl";
import { useParams } from "react-router-dom";
import RolesList from "../../components/roles-and-groups/list/RolesList";
import { useDebounced } from "../../hook/useDebounced";
import { RoleAndGroupsLayout } from "../../RoleAndGroupsLayout";
import { RolesEmptyView } from "../../components/roles-and-groups/empty-view/RolesEmptyView";
import { ReactRouterLinkComponent } from "../../ReactRouterLink";
import { RolesAndGroupsEmptyView } from "../../components/roles-and-groups/empty-view/RolesAndGroupsEmptyView";
import ShowError from "../../ShowError";

const pageSize = 10;
const searchResultSize = -1;

const RolesScreen: React.VFC = function RolesScreen() {
  const { renderToString } = useContext(Context);
  const [searchKeyword, setSearchKeyword] = useState("");
  const { appID } = useParams<{ appID: string }>();

  const isSearch = searchKeyword !== "";
  const [debouncedSearchKey] = useDebounced(searchKeyword, 500);

  const [offset, setOffset] = useState(0);
  // after: is exclusive so if we pass it "offset:0",
  // The first item is excluded.
  // Therefore we have adjust it by -1.
  const cursor = useMemo(() => {
    if (isSearch) {
      // Search always query all rows.
      return undefined;
    }
    return encodeOffsetToCursor(offset);
  }, [isSearch, offset]);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
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

  const { data, loading, error, refetch } = useQuery<
    RolesListQueryQuery,
    RolesListQueryQueryVariables
  >(RolesListQueryDocument, {
    variables: {
      pageSize: isSearch ? searchResultSize : pageSize,
      searchKeyword: debouncedSearchKey,
      cursor,
    },
    fetchPolicy: "network-only",
  });

  const isLoading = loading || data == null;

  const isEmpty = !isLoading && data.roles?.totalCount === 0;

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
    return [{ to: ".", label: <FormattedMessage id="RolesScreen.title" /> }];
  }, []);

  const headerSubItem = useMemo(() => {
    return !isEmpty ? (
      <ReactRouterLinkComponent
        component={RolesAndGroupsEmptyView.CreateButton}
        to={`/project/${appID}/user-management/roles/add-role`}
        text={<FormattedMessage id="RolesEmptyView.button.text" />}
      />
    ) : null;
  }, [appID, isEmpty]);

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <RoleAndGroupsLayout
      headerBreadcrumbs={items}
      headerSubitem={headerSubItem}
    >
      {!isEmpty ? <SearchBox {...searchBoxProps} /> : null}
      {isEmpty ? (
        <RolesEmptyView className={styles.emptyStateContainer} />
      ) : (
        <RolesList
          className={styles.list}
          isSearch={isSearch}
          loading={isLoading}
          offset={offset}
          pageSize={pageSize}
          roles={data?.roles ?? null}
          totalCount={data?.roles?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
        />
      )}
    </RoleAndGroupsLayout>
  );
};

export default RolesScreen;
