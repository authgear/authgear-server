import { useQuery } from "@apollo/client";
import React, { useMemo, useState } from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import {
  RolesListQueryDocument,
  RolesListQueryQuery,
  RolesListQueryQueryVariables,
} from "./query/rolesListQuery.generated";
import styles from "./RolesScreen.module.css";
import { encodeOffsetToCursor } from "../../util/pagination";
import ScreenContent from "../../ScreenContent";
import NavBreadcrumb from "../../NavBreadcrumb";
import { FormattedMessage } from "@oursky/react-messageformat";
import iconBadge from "../../images/badge.svg";
import PrimaryButton from "../../PrimaryButton";
import { useLocation } from "react-router-dom";
import RolesList from "./RolesList";

const pageSize = 10;

interface RolesScreenEmptyStateProps {
  className?: string;
}

const RolesScreenEmptyState: React.VFC<RolesScreenEmptyStateProps> =
  function RolesScreenEmptyState(props) {
    const { className } = props;
    const location = useLocation();
    return (
      <div className={cn(className, styles.emptyStateContainer)}>
        <img className={styles.emptyStateIcon} src={iconBadge} />
        <Text className={styles.emptyStateTitle}>
          <FormattedMessage id="RolesScreen.empty-state.title" />
        </Text>
        <Text className={styles.emptyStateSubtitle}>
          <FormattedMessage id="RolesScreen.empty-state.subtitle" />
        </Text>
        <PrimaryButton
          href={`${location.pathname}/add-role`}
          className={styles.emptyStateButton}
          text={<FormattedMessage id={"RolesScreen.empty-state.button"} />}
          iconProps={{ iconName: "Add" }}
        />
      </div>
    );
  };

const RolesScreen: React.VFC = function RolesScreen() {
  const [searchKeyword, _setSearchKeyword] = useState("");

  const isSearch = searchKeyword !== "";

  const [offset, _setOffset] = useState(0);
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
  const { data, loading } = useQuery<
    RolesListQueryQuery,
    RolesListQueryQueryVariables
  >(RolesListQueryDocument, {
    variables: {
      searchKeyword,
      pageSize,
      cursor,
    },
    fetchPolicy: "network-only",
  });

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="RolesScreen.title" /> }];
  }, []);

  const isEmpty = !loading && data?.roles?.edges?.length === 0;

  return (
    <ScreenContent className={styles.content} layout="list">
      <div className={styles.widget}>
        <NavBreadcrumb className="block" items={items} />
      </div>
      {isEmpty ? (
        <RolesScreenEmptyState className={styles.widget} />
      ) : (
        <RolesList
          className={styles.widget}
          isSearch={isSearch}
          loading={loading}
          offset={offset}
          pageSize={pageSize}
          roles={data?.roles ?? null}
        />
      )}
    </ScreenContent>
  );
};

export default RolesScreen;
