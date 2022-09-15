import React, {
  useMemo,
  useContext,
  useCallback,
  ReactElement,
  ReactNode,
} from "react";
import { DetailsList, SelectionMode, Checkbox, IColumn } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
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

function renderAriaLabel() {
  return "";
}

function LocalCheckbox(props: LocalCheckboxProps): ReactElement {
  const { item, onChangeChecked } = props;

  const onChange = useCallback(
    (_event, checked?: boolean) => {
      onChangeChecked(item.key, checked ?? false);
    },
    [item.key, onChangeChecked]
  );

  return (
    <Checkbox
      checked={item.checked}
      onChange={onChange}
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
  const { renderToString } = useContext(Context);

  const columns: IColumn[] = useMemo(() => {
    return [
      {
        key: "checked",
        fieldName: "checked",
        name: checkedColumnLabel,
        className: styles.cell,
        minWidth: 64,
        maxWidth: 64,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: PriorityListItem) => {
          return (
            <LocalCheckbox item={item} onChangeChecked={onChangeChecked} />
          );
        },
      },
      {
        key: "key",
        fieldName: "key",
        name: keyColumnLabel,
        className: styles.cell,
        minWidth: 0,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: PriorityListItem) => {
          return item.content;
        },
      },
      {
        key: "order",
        name: renderToString("PriorityList.order"),
        className: styles.cell,
        minWidth: 100,
        maxWidth: 100,
        styles: {
          cellTitle: {
            // To align the column title with the order button visually.
            marginLeft: "6px",
          },
        },
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: PriorityListItem, index?: number) => {
          return (
            <OrderButtons
              disabled={item.disabled}
              index={index}
              itemCount={items.length}
              onSwapClicked={onSwap}
              renderAriaLabel={renderAriaLabel}
            />
          );
        },
      },
    ];
  }, [
    checkedColumnLabel,
    keyColumnLabel,
    renderToString,
    items.length,
    onChangeChecked,
    onSwap,
  ]);

  return (
    <DetailsList
      className={className}
      items={items}
      columns={columns}
      selectionMode={SelectionMode.none}
    />
  );
}

export default PriorityList;
