import React, { useContext, useMemo, useCallback, useEffect } from "react";
import cn from "classnames";
import {
  CaretSortIcon,
  CaretUpIcon,
  CaretDownIcon,
  ChevronRightIcon,
} from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";
import { Context, FormattedMessage, Values } from "../../intl";
import Link from "../../Link";
import PaginationWidget from "../../PaginationWidget";
import {
  AuditLogListFragment,
  AuditLogEdgesNodeFragment,
} from "./query/auditLogListQuery.generated";
import { SortDirection } from "./globalTypes.generated";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import { useDebounced } from "../../hook/useDebounced";

import styles from "./AuditLogList.module.css";
import { useParams } from "react-router-dom";

const PLACEHOLDER = "-";
const SHIMMER_ROW_COUNT = 5;

export interface AuditLogListProps {
  className?: string;
  loading: boolean;
  auditLogs: AuditLogListFragment | null;
  searchParams: string;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
  onToggleSortDirection?: () => void;
  sortDirection?: SortDirection;
}

interface AuditLogListItem {
  id: string;
  activityType: string;
  createdAt: string;
  userID: string | null;
  userPrimary: string;
  userSecondary: string | null;
}

function getUserDisplayFromAuditLog(
  renderToString: (id: string, values: Values | undefined) => string,
  node: AuditLogEdgesNodeFragment
): { primary: string; secondary: string | null } {
  const user = node.user;
  if (user != null) {
    const name = user.formattedName?.trim() || null;
    const rawID = extractRawID(user.id);
    const primary = name ?? user.endUserAccountID ?? rawID;
    return {
      primary,
      secondary: primary === rawID ? null : rawID,
    };
  }

  // No linked user record; fall back to the raw id from the event payload.
  const rawUserID = node.data?.payload?.user?.id;
  if (rawUserID != null) {
    return {
      primary: renderToString("AuditLogList.label.user-id", { id: rawUserID }),
      secondary: null,
    };
  }

  return { primary: PLACEHOLDER, secondary: null };
}

function UserCellContent(props: {
  primary: string;
  secondary: string | null;
}): React.ReactElement {
  const { primary, secondary } = props;
  return (
    <div className={styles.userCell}>
      <Text size="2" className={cn(styles.cellText, styles.userPrimary)}>
        {primary}
      </Text>
      {secondary != null ? (
        <Text
          size="1"
          color="gray"
          className={cn(styles.cellText, styles.userSecondary)}
        >
          {secondary}
        </Text>
      ) : null}
    </div>
  );
}

const AuditLogList: React.VFC<AuditLogListProps> = function AuditLogList(
  props
) {
  const {
    className,
    loading: rawLoading,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
    onToggleSortDirection,
    sortDirection,
  } = props;
  const edges = props.auditLogs?.edges;
  const searchParams = props.searchParams;

  const { appID } = useParams() as { appID: string };

  const [loading] = useDebounced(rawLoading, 500);

  const { renderToString, locale } = useContext(Context);

  const items: AuditLogListItem[] = useMemo(() => {
    const items: AuditLogListItem[] = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          const userID = node.user?.id ?? null;
          const userDisplay = getUserDisplayFromAuditLog(renderToString, node);
          items.push({
            id: node.id,
            userID,
            userPrimary: userDisplay.primary,
            userSecondary: userDisplay.secondary,
            createdAt: formatDatetime(locale, node.createdAt)!,
            activityType: renderToString(
              "AuditLogActivityType." + node.activityType
            ),
          });
        }
      }
    }
    return items;
  }, [edges, locale, renderToString]);

  const listWrapperRef = React.useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    listWrapperRef.current?.scrollTo(0, 0);
  }, [items]);

  const onClickSortHeader = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      onToggleSortDirection?.();
      onChangeOffset?.(0);
    },
    [onToggleSortDirection, onChangeOffset]
  );

  const isEmpty = !loading && items.length === 0;

  const SortIcon =
    sortDirection === SortDirection.Asc
      ? CaretUpIcon
      : sortDirection === SortDirection.Desc
      ? CaretDownIcon
      : CaretSortIcon;

  return (
    <div className={cn(styles.root, className)}>
      <div
        ref={listWrapperRef}
        className={styles.tableWrapper}
        data-is-scrollable="true"
      >
        <div className={styles.table}>
          {/* Header */}
          <div className={styles.tableHeader}>
            <div className={styles.headerCellUser}>
              <FormattedMessage id="AuditLogList.column.user-id" />
            </div>
            <div className={styles.headerCellActivityType}>
              <FormattedMessage id="AuditLogList.column.activity-type" />
            </div>
            <div className={styles.headerCellCreatedAt}>
              <button
                className={styles.sortButton}
                onClick={onClickSortHeader}
                type="button"
              >
                <FormattedMessage id="AuditLogList.column.created-at" />
                <SortIcon className={styles.sortIcon} />
              </button>
            </div>
            <div className={styles.headerCellChevron} aria-hidden={true} />
          </div>

          {/* Shimmer rows while loading */}
          {loading
            ? Array.from({ length: SHIMMER_ROW_COUNT }).map((_, i) => (
                <div key={i} className={styles.shimmerRow}>
                  <div className={cn(styles.shimmerCell, styles.shimmerUser)} />
                  <div
                    className={cn(
                      styles.shimmerCell,
                      styles.shimmerActivityType
                    )}
                  />
                  <div
                    className={cn(styles.shimmerCell, styles.shimmerCreatedAt)}
                  />
                  <div
                    className={cn(styles.shimmerCell, styles.shimmerChevron)}
                  />
                </div>
              ))
            : items.map((item) => {
                const detailHref = `/project/${appID}/audit-log/${item.id}/details`;
                const detailState: any = { searchParams };
                const userHref =
                  item.userID != null
                    ? `/project/${appID}/user-management/users/${item.userID}/details#logs`
                    : null;

                return (
                  <div key={item.id} className={styles.tableRow}>
                    <div className={styles.cellUser}>
                      {userHref != null ? (
                        <Link className={styles.cellUserLink} to={userHref}>
                          <UserCellContent
                            primary={item.userPrimary}
                            secondary={item.userSecondary}
                          />
                        </Link>
                      ) : (
                        <UserCellContent
                          primary={item.userPrimary}
                          secondary={item.userSecondary}
                        />
                      )}
                    </div>
                    <div className={styles.cellActivityType}>
                      <Link
                        className={styles.activityTypeLink}
                        to={detailHref}
                        state={detailState}
                      >
                        <Text size="2">{item.activityType}</Text>
                      </Link>
                    </div>
                    <div className={styles.cellCreatedAt}>
                      <Text size="2" className={styles.cellText}>
                        {item.createdAt}
                      </Text>
                    </div>
                    <div className={styles.cellChevron}>
                      <Link to={detailHref} state={detailState}>
                        <ChevronRightIcon
                          className={styles.chevronIcon}
                          width="1rem"
                          height="1rem"
                        />
                      </Link>
                    </div>
                  </div>
                );
              })}

          {/* Empty state row inside the table */}
          {isEmpty ? (
            <div className={styles.emptyRow}>
              <Text size="2" className={styles.emptyText}>
                <FormattedMessage id="AuditLogList.empty" />
              </Text>
            </div>
          ) : null}
        </div>
      </div>

      <PaginationWidget
        className={cn(styles.pagination, isEmpty && styles.paginationHidden)}
        offset={offset}
        pageSize={pageSize}
        totalCount={totalCount}
        onChangeOffset={onChangeOffset}
      />
    </div>
  );
};

export default AuditLogList;
