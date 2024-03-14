import React from "react";
import {
  DetailsListLayoutMode,
  IColumn,
  IDetailsRowProps,
  IRenderFunction,
  SelectionMode,
  ShimmeredDetailsList,
} from "@fluentui/react";
import styles from "./RolesAndGroupsBaseList.module.css";
import cn from "classnames";
import PaginationWidget from "../../../PaginationWidget";

interface RolesAndGroupsBaseListProps<T> {
  className?: string;
  loading: boolean;
  isSearch: boolean;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;

  onRenderRow: IRenderFunction<IDetailsRowProps>;
  onRenderItemColumn: (
    item: T,
    index?: number,
    column?: IColumn
  ) => React.ReactNode;
  items: T[];
  columns: IColumn[];
}

function RolesAndGroupsBaseList<T>(
  props: RolesAndGroupsBaseListProps<T>
): React.ReactElement {
  const {
    className,
    loading,
    isSearch,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
    onRenderRow,
    onRenderItemColumn,
    items,
    columns,
  } = props;

  return (
    <>
      <div className={cn(styles.listWrapper, className)}>
        <ShimmeredDetailsList
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
      {!isSearch ? (
        <PaginationWidget
          className={cn(styles.pagination)}
          offset={offset}
          pageSize={pageSize}
          totalCount={totalCount}
          onChangeOffset={onChangeOffset}
        />
      ) : null}
    </>
  );
}

export default RolesAndGroupsBaseList;
