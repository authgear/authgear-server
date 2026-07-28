import React from "react";
import { IColumn } from "@fluentui/react";
import { Spinner, Table, Text } from "@radix-ui/themes";
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

  onRenderItemColumn: (
    item: T,
    index?: number,
    column?: IColumn
  ) => React.ReactNode;
  items: T[];
  columns: IColumn[];
  emptyText: string;
  onItemClick?: (item: T) => void;
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
                  key={index}
                  className={onItemClick != null ? styles.clickableRow : undefined}
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
