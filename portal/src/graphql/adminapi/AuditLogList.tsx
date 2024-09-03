import React, { useContext, useMemo, useCallback, useEffect } from "react";
import cn from "classnames";
import {
  IColumn,
  ColumnActionsMode,
  SelectionMode,
  DetailsListLayoutMode,
  ShimmeredDetailsList,
  MessageBar,
} from "@fluentui/react";
import { Context, FormattedMessage, Values } from "@oursky/react-messageformat";
import Link from "../../Link";
import PaginationWidget from "../../PaginationWidget";
import {
  AuditLogListFragment,
  AuditLogEdgesNodeFragment,
} from "./query/auditLogListQuery.generated";
import { SortDirection } from "./globalTypes.generated";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import useDelayedValue from "../../hook/useDelayedValue";

import styles from "./AuditLogList.module.css";
import { useParams } from "react-router-dom";

const PLACEHOLDER = "-";

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
  rawUserID: string | null;
}

function getRawUserIDFromAuditLog(
  renderToString: (id: string, values: Values | undefined) => string,
  node: AuditLogEdgesNodeFragment
): string | null {
  // The simple case is just use the user.id.
  const userID = node.user?.id ?? null;
  if (userID != null) {
    return extractRawID(userID);
  }

  // Otherwise use the user ID in the payload.
  const rawUserID = node.data?.payload?.user?.id;
  if (rawUserID != null) {
    return renderToString("AuditLogList.label.user-id", {
      id: rawUserID,
    });
  }

  return null;
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

  const loading = useDelayedValue(rawLoading, 500);

  const { renderToString, locale } = useContext(Context);

  const columns: IColumn[] = useMemo(
    () => [
      {
        key: "activityType",
        fieldName: "activityType",
        name: renderToString("AuditLogList.column.activity-type"),
        maxWidth: 300,
        minWidth: 300,
        columnActionsMode: ColumnActionsMode.disabled,
      },
      {
        key: "createdAt",
        fieldName: "createdAt",
        name: renderToString("AuditLogList.column.created-at"),
        maxWidth: 220,
        minWidth: 220,
        isSorted: true,
        isSortedDescending: sortDirection === SortDirection.Desc,
        iconName: "SortLines",
        iconClassName: styles.sortIcon,
      },
      {
        key: "rawUserID",
        fieldName: "rawUserID",
        name: renderToString("AuditLogList.column.user-id"),
        minWidth: 430,
        columnActionsMode: ColumnActionsMode.disabled,
      },
    ],
    [renderToString, sortDirection]
  );

  const items: AuditLogListItem[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          const userID = node.user?.id ?? null;
          const rawUserID = getRawUserIDFromAuditLog(renderToString, node);
          items.push({
            id: node.id,
            userID,
            rawUserID,
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

  // Reset scroll position when items change.
  const listWrapperRef = React.useRef<HTMLDivElement | null>(null);
  useEffect(() => {
    listWrapperRef.current?.scrollTo(0, 0);
  }, [items]);

  const onRenderItemColumn = useCallback(
    (item: AuditLogListItem, _index?: number, column?: IColumn) => {
      const text = item[column?.key as keyof AuditLogListItem] ?? PLACEHOLDER;

      let href: string | null = null;
      const state: any = {};
      switch (column?.key) {
        case "activityType":
          href = `/project/${appID}/audit-log/${item.id}/details`;
          state["searchParams"] = searchParams;
          break;
        case "rawUserID":
          if (item.userID != null) {
            href = `/project/${appID}/users/${item.userID}/details`;
          }
          break;
        default:
          break;
      }

      if (href != null) {
        return (
          <Link to={href} state={state}>
            {text}
          </Link>
        );
      }
      return <span>{text}</span>;
    },
    [appID, searchParams]
  );

  const onColumnHeaderClick = useCallback(
    (_e, column) => {
      if (column != null) {
        if (column.key === "createdAt") {
          onToggleSortDirection?.();
          onChangeOffset?.(0);
        }
      }
    },
    [onToggleSortDirection, onChangeOffset]
  );

  const isEmpty = !loading && items.length === 0;

  return (
    <>
      <div className={cn(styles.root, className)}>
        <div
          ref={listWrapperRef}
          className={cn(styles.listWrapper, isEmpty && styles.empty)}
          data-is-scrollable="true"
        >
          <ShimmeredDetailsList
            className={styles.list}
            enableShimmer={loading}
            enableUpdateAnimations={false}
            selectionMode={SelectionMode.none}
            layoutMode={DetailsListLayoutMode.justified}
            onColumnHeaderClick={onColumnHeaderClick}
            onRenderItemColumn={onRenderItemColumn}
            columns={columns}
            items={items}
          />
        </div>
        <PaginationWidget
          className={cn(styles.pagination, isEmpty && styles.empty)}
          offset={offset}
          pageSize={pageSize}
          totalCount={totalCount}
          onChangeOffset={onChangeOffset}
        />
        {isEmpty ? (
          <MessageBar>
            <FormattedMessage id="AuditLogList.empty" />
          </MessageBar>
        ) : null}
      </div>
    </>
  );
};

export default AuditLogList;
