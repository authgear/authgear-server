import React from "react";
import {
  DetailsList,
  IDetailsListProps,
  IColumn,
  IconButton,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";

import styles from "./DetailsListWithOrdering.module.scss";

interface DetailsListWithOrderingProps extends IDetailsListProps {
  onSwapClicked: (index1: number, index2: number) => void;
  columns: IColumn[];
  onRenderItemColumn: (
    item?: any,
    index?: number,
    column?: IColumn
  ) => React.ReactNode;
  orderColumnMinWidth?: number;
  orderColumnMaxWidth?: number;
}

interface OrderColumnButtonsProps {
  index?: number;
  onSwapClicked: (index1: number, index2: number) => void;
}

const OrderColumnButtons: React.FC<OrderColumnButtonsProps> = function OrderColumnButtons(
  props: OrderColumnButtonsProps
) {
  const onUpClicked = React.useCallback(() => {
    if (props.index == null) {
      return;
    }
    props.onSwapClicked(props.index, props.index - 1);
  }, [props]);
  const onDownClicked = React.useCallback(() => {
    if (props.index == null) {
      return;
    }
    props.onSwapClicked(props.index, props.index + 1);
  }, [props]);
  return (
    <div>
      <IconButton
        className={styles.orderColumnButton}
        onClick={onDownClicked}
        iconProps={{ iconName: "ChevronDown" }}
      />
      <IconButton
        className={styles.orderColumnButton}
        onClick={onUpClicked}
        iconProps={{ iconName: "ChevronUp" }}
      />
    </div>
  );
};

export function swap<T>(items: T[], index1: number, index2: number): T[] {
  const newItems = [...items];
  const thisItem = newItems[index1];
  const thatItem = newItems[index2];
  if (
    index1 < 0 ||
    index2 < 0 ||
    index1 >= items.length ||
    index2 >= items.length
  ) {
    return items;
  }
  newItems[index1] = thatItem;
  newItems[index2] = thisItem;
  return newItems;
}

const DetailsListWithOrdering: React.FC<DetailsListWithOrderingProps> = function DetailsListWithOrdering(
  props: DetailsListWithOrderingProps
) {
  const { renderToString } = React.useContext(Context);
  const onRenderItemColumn = React.useCallback(
    (item?: any, index?: number, column?: IColumn) => {
      if (column?.key === "order") {
        return (
          <OrderColumnButtons
            index={index}
            onSwapClicked={props.onSwapClicked}
          />
        );
      }
      return props.onRenderItemColumn(item, index, column);
    },
    [props]
  );

  const columns: IColumn[] = [
    ...props.columns,
    {
      key: "order",
      fieldName: "order",
      name: renderToString("DetailsListWithOrdering.order"),
      minWidth: props.orderColumnMinWidth ?? 100,
      maxWidth: props.orderColumnMaxWidth ?? 100,
    },
  ];

  return (
    <DetailsList
      {...props}
      columns={columns}
      onRenderItemColumn={onRenderItemColumn}
    />
  );
};

export default DetailsListWithOrdering;
