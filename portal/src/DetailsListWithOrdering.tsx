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
  renderAriaLabel: (index?: number) => string;
  orderColumnMinWidth?: number;
  orderColumnMaxWidth?: number;
}

interface OrderColumnButtonsProps {
  index?: number;
  itemCount: number;
  onSwapClicked: (index1: number, index2: number) => void;
  renderAriaLabel: (index?: number) => string;
}

const OrderColumnButtons: React.FC<OrderColumnButtonsProps> = function OrderColumnButtons(
  props: OrderColumnButtonsProps
) {
  const { index, itemCount, onSwapClicked, renderAriaLabel } = props;
  const { renderToString } = React.useContext(Context);
  const onUpClicked = React.useCallback(() => {
    if (index == null) {
      return;
    }
    onSwapClicked(index, index - 1);
  }, [index, onSwapClicked]);
  const onDownClicked = React.useCallback(() => {
    if (index == null) {
      return;
    }
    onSwapClicked(index, index + 1);
  }, [index, onSwapClicked]);

  const ariaLabelUp = React.useMemo(() => {
    return [renderAriaLabel(index), renderToString("up")].join(" | ");
  }, [renderAriaLabel, index, renderToString]);
  const ariaLabelDown = React.useMemo(() => {
    return [renderAriaLabel(index), renderToString("down")].join(" | ");
  }, [renderAriaLabel, index, renderToString]);

  return (
    <div>
      <IconButton
        className={styles.orderColumnButton}
        disabled={index === itemCount - 1}
        onClick={onDownClicked}
        iconProps={{ iconName: "ChevronDown" }}
        ariaLabel={ariaLabelDown}
      />
      <IconButton
        className={styles.orderColumnButton}
        disabled={index === 0}
        onClick={onUpClicked}
        iconProps={{ iconName: "ChevronUp" }}
        ariaLabel={ariaLabelUp}
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
            itemCount={props.items.length}
            onSwapClicked={props.onSwapClicked}
            renderAriaLabel={props.renderAriaLabel}
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
