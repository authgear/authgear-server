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
import { Context } from "@oursky/react-messageformat";
import ReactRouterLink from "../../ReactRouterLink";
import PaginationWidget from "../../PaginationWidget";
import { AuditLogListQuery_auditLogs } from "./__generated__/AuditLogListQuery";
import { formatDatetime } from "../../util/formatDatetime";
import { extractRawID } from "../../util/graphql";
import useDelayedValue from "../../hook/useDelayedValue";

import styles from "./AuditLogList.module.scss";

const PLACEHOLDER = "-";

export interface AuditLogListProps {
  className?: string;
  loading: boolean;
  auditLogs: AuditLogListQuery_auditLogs | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface AuditLogListItem {
  id: string;
  activityType: string;
  createdAt: string;
  userID: string | null;
  rawUserID: string | null;
}

const AuditLogList: React.FC<AuditLogListProps> = function AuditLogList(props) {
  const {
    className,
    loading: rawLoading,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  } = props;
  const edges = props.auditLogs?.edges;

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
      columnActionsMode: ColumnActionsMode.disabled,
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
          const rawUserID = userID ? extractRawID(userID) : null;
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
      const href = `../users/${item.userID}/details/`;
      const text = item[column?.key as keyof AuditLogListItem] ?? PLACEHOLDER;
      switch (column?.key) {
        case "rawUserID":
          return (
            <ReactRouterLink to={href} component={FluentLink}>
              {text}
            </ReactRouterLink>
          );
        default:
          return <span>{text}</span>;
      }
    },
    []
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
