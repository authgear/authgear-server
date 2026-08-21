import React from "react";
import cn from "classnames";
import styles from "./CardTable.module.css";

export interface CardTableProps {
  className?: string;
  children?: React.ReactNode;
}

// A bordered, rounded "card" table built from divs (matches ScopeList /
// ApplicationList). Compose it as:
//   <CardTable>
//     <CardTable.Header>
//       <CardTable.HeaderCell className={styles.colName}>Name</CardTable.HeaderCell>
//       ...
//     </CardTable.Header>
//     {items.map((item) => (
//       <CardTable.Row key={item.id}>
//         <CardTable.Cell className={styles.colName}>{item.name}</CardTable.Cell>
//         ...
//       </CardTable.Row>
//     ))}
//   </CardTable>
// The component owns the frame, header/row backgrounds, separators and cell
// padding; callers supply per-column width/flex via className.
function CardTableRoot({
  className,
  children,
}: CardTableProps): React.ReactElement {
  return (
    <div className={cn(styles.wrapper, className)}>
      <div className={styles.table}>{children}</div>
    </div>
  );
}

function CardTableHeader({
  className,
  children,
}: CardTableProps): React.ReactElement {
  return <div className={cn(styles.header, className)}>{children}</div>;
}

export interface CardTableRowProps
  extends React.HTMLAttributes<HTMLDivElement> {}

function CardTableRow({
  className,
  children,
  ...rest
}: CardTableRowProps): React.ReactElement {
  return (
    <div className={cn(styles.row, className)} {...rest}>
      {children}
    </div>
  );
}

export interface CardTableCellProps
  extends React.HTMLAttributes<HTMLDivElement> {}

function CardTableHeaderCell({
  className,
  children,
  ...rest
}: CardTableCellProps): React.ReactElement {
  return (
    <div className={cn(styles.headerCell, className)} {...rest}>
      {children}
    </div>
  );
}

function CardTableCell({
  className,
  children,
  ...rest
}: CardTableCellProps): React.ReactElement {
  return (
    <div className={cn(styles.cell, className)} {...rest}>
      {children}
    </div>
  );
}

export const CardTable = Object.assign(CardTableRoot, {
  Header: CardTableHeader,
  HeaderCell: CardTableHeaderCell,
  Row: CardTableRow,
  Cell: CardTableCell,
});
