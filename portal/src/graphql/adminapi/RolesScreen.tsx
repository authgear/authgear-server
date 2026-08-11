import { useQuery } from "@apollo/client";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { Spinner, Text } from "@radix-ui/themes";
import { Cross2Icon } from "@radix-ui/react-icons";
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
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
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

  const onChangeSearchKeyword = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchKeyword(e.target.value);
      // Reset offset when search keyword was changed.
      setOffset(0);
    },
    []
  );

  const onClearSearchKeyword = useCallback(() => {
    setSearchKeyword("");
    // Reset offset when search keyword was changed.
    setOffset(0);
  }, []);

  const { data, previousData, loading, error, refetch } = useQuery<
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

  const currentData = data ?? previousData;
  // Only the very first load has no data at all; subsequent refetches keep
  // showing the previous data so the chrome does not flicker.
  const isInitialLoading = loading && currentData == null;

  const isEmpty = !loading && currentData?.roles?.totalCount === 0;

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="RolesScreen.title" /> }];
  }, []);

  const headerSubItem = useMemo(() => {
    return !isInitialLoading && !isEmpty ? (
      <ReactRouterLinkComponent
        component={RolesAndGroupsEmptyView.CreateButton}
        to={`/project/${appID}/user-management/roles/add-role`}
        text={<FormattedMessage id="RolesEmptyView.button.text" />}
      />
    ) : null;
  }, [appID, isInitialLoading, isEmpty]);

  if (error != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <RoleAndGroupsLayout
      headerBreadcrumbs={items}
      headerSubitem={headerSubItem}
      headerDescription={
        !isInitialLoading && !isEmpty ? (
          <Text as="p" size="2" color="gray">
            <FormattedMessage id="RolesScreen.description" />
          </Text>
        ) : null
      }
    >
      {!isInitialLoading && !isEmpty ? (
        <div className={styles.searchField}>
          <TextField
            size="2"
            type="search"
            value={searchKeyword}
            placeholder={renderToString("search")}
            iconStart={TextFieldIcon.MagnifyingGlass}
            onChange={onChangeSearchKeyword}
            suffixPlain={true}
            suffix={
              searchKeyword !== "" ? (
                <button
                  type="button"
                  className={styles.searchClearButton}
                  aria-label={renderToString("APIResourcesScreen.clear-search")}
                  onClick={onClearSearchKeyword}
                >
                  <Cross2Icon className={styles.searchClearIcon} />
                </button>
              ) : undefined
            }
          />
        </div>
      ) : null}
      {isInitialLoading ? (
        <div className={styles.loadingContainer}>
          <Spinner size="3" />
        </div>
      ) : isEmpty ? (
        <RolesEmptyView className={styles.emptyStateContainer} />
      ) : (
        <RolesList
          className={styles.list}
          isSearch={isSearch}
          loading={loading}
          offset={offset}
          pageSize={pageSize}
          roles={currentData?.roles ?? null}
          totalCount={currentData?.roles?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
        />
      )}
    </RoleAndGroupsLayout>
  );
};

export default RolesScreen;
