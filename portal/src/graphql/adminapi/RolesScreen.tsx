import { useQuery } from "@apollo/client";
import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { ISearchBoxProps, SearchBox, MessageBar } from "@fluentui/react";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "./query/rolesListQuery.generated";
import styles from "./RolesScreen.module.css";
import { encodeOffsetToCursor } from "../../util/pagination";
import ScreenContent from "../../ScreenContent";
import NavBreadcrumb from "../../NavBreadcrumb";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import iconBadge from "../../images/badge.svg";
import PrimaryButton from "../../PrimaryButton";
import { useLocation, useNavigate } from "react-router-dom";
import RolesList from "./RolesList";
import useDelayedValue from "../../hook/useDelayedValue";
import { RolesAndGroupsEmptyView } from "../../components/roles-and-groups/RolesAndGroupsEmptyView";

const pageSize = 10;
const searchResultSize = -1;

interface CreateRoleButtonProps {
  className?: string;
}

const CreateRoleButton: React.VFC<CreateRoleButtonProps> =
  function CreateRoleButton(props) {
    const { className } = props;
    const navigate = useNavigate();
    return (
      <PrimaryButton
        className={className}
        text={<FormattedMessage id={"RolesScreen.empty-state.button"} />}
        iconProps={{ iconName: "Add" }}
        onClick={(e: React.MouseEvent<unknown>) => {
          e.preventDefault();
          e.stopPropagation();
          navigate("./add-role");
        }}
      />
    );
  };

interface RolesScreenEmptyStateProps {
  className?: string;
}

const RolesScreenEmptyState: React.VFC<RolesScreenEmptyStateProps> =
  function RolesScreenEmptyState(props) {
    const { className } = props;
    const location = useLocation();
    return (
      <RolesAndGroupsEmptyView
        className={cn(className, styles.emptyStateContainer)}
        icon={<img src={iconBadge} />}
        title={<FormattedMessage id="RolesScreen.empty-state.title" />}
        description={<FormattedMessage id="RolesScreen.empty-state.subtitle" />}
        button={
          <RolesAndGroupsEmptyView.CreateButton
            href={`${location.pathname}/add-role`}
            text={<FormattedMessage id="RolesScreen.empty-state.button" />}
          />
        }
      />
    );
  };

// eslint-disable-next-line complexity
const RolesScreen: React.VFC = function RolesScreen() {
  const { renderToString } = useContext(Context);
  const [searchKeyword, setSearchKeyword] = useState("");

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

  const { data, loading, previousData } = useQuery<
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

  const isInitialLoading = loading && previousData == null;

  const isEmpty = !isInitialLoading && data?.roles?.totalCount === 0;
  const isSearchEmpty = isSearch && data?.roles?.edges?.length === 0;

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

  return (
    <ScreenContent className={styles.content} layout="list">
      <div className={styles.widget}>
        <div className={styles.titleContainer}>
          <NavBreadcrumb className="block" items={items} />
          {!isEmpty ? <CreateRoleButton /> : null}
        </div>
        {!isEmpty ? <SearchBox {...searchBoxProps} /> : null}
      </div>
      {isEmpty ? (
        <RolesScreenEmptyState className={styles.widget} />
      ) : isSearchEmpty ? (
        <MessageBar className={cn(styles.widget, styles.message)}>
          <FormattedMessage id="RolesScreen.empty.search" />
        </MessageBar>
      ) : (
        <RolesList
          className={styles.widget}
          isSearch={isSearch}
          loading={isInitialLoading}
          offset={offset}
          pageSize={pageSize}
          roles={data?.roles ?? null}
          totalCount={data?.roles?.totalCount ?? undefined}
          onChangeOffset={onChangeOffset}
        />
      )}
    </ScreenContent>
  );
};

export default RolesScreen;
