import React from "react";
import { Spinner, Table, Text } from "@radix-ui/themes";
import styles from "./RolesAndGroupsBaseList.module.css";
import cn from "classnames";
import PaginationWidget from "../../../../PaginationWidget";

export interface RolesAndGroupsListColumn {
  key: string;
  name: string;
  fieldName?: string;
  minWidth?: number;
  maxWidth?: number;
  isResizable?: boolean;
}

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

  onRenderItemColumn: (
    item: T,
    index?: number,
    column?: RolesAndGroupsListColumn
  ) => React.ReactNode;
  items: T[];
  columns: RolesAndGroupsListColumn[];
  emptyText: string;
  onItemClick?: (item: T) => void;
  // Stable per-item key. Falls back to the array index, which is only correct
  // while the whole page is replaced at once.
  getItemKey?: (item: T, index: number) => React.Key;
}

function RolesAndGroupsBaseList<T>(
  props: RolesAndGroupsBaseListProps<T>
): React.ReactElement {
  const {
    className,
    loading,
    pagination,
    onRenderItemColumn,
    items,
    columns,
    emptyText,
    onItemClick,
    getItemKey,
  } = props;

  const isEmpty = items.length === 0 && !loading;

  return isEmpty ? (
    <Text as="p" size="2" color="gray" className={styles.message}>
      {emptyText}
    </Text>
  ) : (
    <>
      <div className={cn(styles.listWrapper, className)}>
        {loading ? (
          <div className={styles.loading}>
            <Spinner />
          </div>
        ) : (
          <Table.Root variant="surface">
            <Table.Header>
              <Table.Row>
                {columns.map((column) => (
                  <Table.ColumnHeaderCell key={column.key}>
                    {column.name}
                  </Table.ColumnHeaderCell>
                ))}
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {items.map((item, index) => (
                <Table.Row
                  key={getItemKey?.(item, index) ?? index}
                  className={
                    onItemClick != null ? styles.clickableRow : undefined
                  }
                  onClick={() => onItemClick?.(item)}
                >
                  {columns.map((column) => (
                    <Table.Cell key={column.key}>
                      {onRenderItemColumn(item, index, column)}
                    </Table.Cell>
                  ))}
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        )}
      </div>
      {pagination != null && !pagination.isSearch ? (
        <PaginationWidget className={styles.pagination} {...pagination} />
      ) : null}
    </>
  );
}

export default RolesAndGroupsBaseList;
