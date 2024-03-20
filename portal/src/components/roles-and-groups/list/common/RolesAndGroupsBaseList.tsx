import React, { useCallback } from "react";
import {
  DetailsListLayoutMode,
  IColumn,
  IDetailsRowProps,
  IRenderFunction,
  MessageBar,
  SelectionMode,
  ShimmeredDetailsList,
} from "@fluentui/react";
import styles from "./RolesAndGroupsBaseList.module.css";
import cn from "classnames";
import PaginationWidget from "../../../../PaginationWidget";

interface PaginationProps {
  isSearch: boolean;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface RolesAndGroupsBaseListProps<T> {
  className?: string;
  loading?: boolean;
  pagination?: PaginationProps;

  onRenderRow: IRenderFunction<IDetailsRowProps>;
  onRenderItemColumn: (
    item: T,
    index?: number,
    column?: IColumn
  ) => React.ReactNode;
  items: T[];
  columns: IColumn[];
  emptyText: string;
}

function RolesAndGroupsBaseList<T>(
  props: RolesAndGroupsBaseListProps<T>
): React.ReactElement {
  const {
    className,
    loading,
    pagination,
    onRenderRow,
    onRenderItemColumn,
    items,
    columns,
    emptyText,
  } = props;

  // NOTE: Avoid DetailsList automatically take key column of item as key of react list, causing duplicated key error
  // Ref https://github.com/microsoft/fluentui/blob/19195df9f17f287bf4c66dd30453dd0d0a0ced93/packages/react/src/components/DetailsList/DetailsList.base.tsx#L1499
  const getKey = useCallback((item: any, index?: number) => {
    const baseKey = index?.toString() ?? "0";
    const itemKey = item?.key ?? null;

    return typeof itemKey === "string" ? `${itemKey}-${baseKey}` : baseKey;
  }, []);

  const isEmpty = items.length === 0 && !loading;

  return isEmpty ? (
    <MessageBar className={styles.message}>{emptyText}</MessageBar>
  ) : (
    <>
      <div
        className={cn(styles.listWrapper, className)}
        // For DetailList to correctly know what to display
        // https://developer.microsoft.com/en-us/fluentui#/controls/web/detailslist
        data-is-scrollable="true"
      >
        <ShimmeredDetailsList
          getKey={getKey}
          enableShimmer={loading}
          enableUpdateAnimations={false}
          onRenderRow={onRenderRow}
          onRenderItemColumn={onRenderItemColumn}
          selectionMode={SelectionMode.none}
          layoutMode={DetailsListLayoutMode.justified}
          items={items}
          columns={columns}
        />
      </div>
      {pagination != null && !pagination.isSearch ? (
        <PaginationWidget className={styles.pagination} {...pagination} />
      ) : null}
    </>
  );
}

export default RolesAndGroupsBaseList;
