import React, { useCallback, ReactElement, ReactNode } from "react";
import cn from "classnames";
import { Checkbox } from "@radix-ui/themes";
import { FormattedMessage } from "./intl";
import styles from "./PriorityList.module.css";
import OrderButtons from "./OrderButtons";

export interface PriorityListItem {
  key: string;
  checked: boolean;
  disabled: boolean;
  content: ReactNode;
}

export interface PriorityListProps {
  className?: string;
  items: PriorityListItem[];
  checkedColumnLabel: string;
  keyColumnLabel: string;
  onChangeChecked: (key: string, checked: boolean) => void;
  onSwap: (index1: number, index2: number) => void;
}

interface LocalCheckboxProps {
  item: PriorityListItem;
  onChangeChecked: (key: string, checked: boolean) => void;
}

function LocalCheckbox(props: LocalCheckboxProps): ReactElement {
  const { item, onChangeChecked } = props;

  const onCheckedChange = useCallback(
    (checked: boolean | "indeterminate") => {
      if (checked === "indeterminate") {
        return;
      }
      onChangeChecked(item.key, checked);
    },
    [item.key, onChangeChecked]
  );

  return (
    <Checkbox
      checked={item.checked}
      onCheckedChange={onCheckedChange}
      disabled={Boolean(item.disabled && !item.checked)}
    />
  );
}

function PriorityList(props: PriorityListProps): ReactElement {
  const {
    className,
    items,
    checkedColumnLabel,
    keyColumnLabel,
    onChangeChecked,
    onSwap,
  } = props;

  return (
    <div className={cn(styles.tableWrapper, className)}>
      <div className={styles.table}>
        <div className={styles.tableHeader}>
          <div className={styles.headerCellChecked}>{checkedColumnLabel}</div>
          <div className={styles.headerCellKey}>{keyColumnLabel}</div>
          <div className={styles.headerCellOrder}>
            <FormattedMessage id="PriorityList.order" />
          </div>
        </div>
        {items.map((item, index) => (
          <div key={item.key} className={styles.tableRow}>
            <div className={styles.cellChecked}>
              <LocalCheckbox item={item} onChangeChecked={onChangeChecked} />
            </div>
            <div className={styles.cellKey}>{item.content}</div>
            <div className={styles.cellOrder}>
              <OrderButtons
                disabled={item.disabled}
                index={index}
                itemCount={items.length}
                onSwapClicked={onSwap}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default PriorityList;
