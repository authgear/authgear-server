import React, { useContext, useMemo, useCallback } from "react";
import cn from "classnames";
import {
  IColumn,
  ColumnActionsMode,
  SelectionMode,
  DetailsListLayoutMode,
  ShimmeredDetailsList,
  Link as FluentLink,
} from "@fluentui/react";
import { Context, Values } from "@oursky/react-messageformat";
import ReactRouterLink from "../../ReactRouterLink";
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

  const loading = useDelayedValue(rawLoading, 500);

  const { renderToString, locale } = useContext(Context);

  const columns: IColumn[] = [
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
      maxWidth: 150,
      minWidth: 150,
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
  ];

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

  const onRenderItemColumn = useCallback(
    (item: AuditLogListItem, _index?: number, column?: IColumn) => {
      const text = item[column?.key as keyof AuditLogListItem] ?? PLACEHOLDER;

      let href: string | null = null;
      const state: any = {};
      switch (column?.key) {
        case "activityType":
          href = `./${item.id}/details`;
          state["searchParams"] = searchParams;
          break;
        case "rawUserID":
          if (item.userID != null) {
            href = `./../users/${item.userID}/details`;
          }
          break;
        default:
          break;
      }

      if (href != null) {
        return (
          <ReactRouterLink to={href} state={state} component={FluentLink}>
            {text}
          </ReactRouterLink>
        );
      }
      return <span>{text}</span>;
    },
    [searchParams]
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

  return (
    <>
      <div className={cn(styles.root, className)}>
        <div className={styles.listWrapper}>
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
          className={styles.pagination}
          offset={offset}
          pageSize={pageSize}
          totalCount={totalCount}
          onChangeOffset={onChangeOffset}
        />
      </div>
    </>
  );
};

export default AuditLogList;
