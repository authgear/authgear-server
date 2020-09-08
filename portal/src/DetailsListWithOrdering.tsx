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
  onSwapClicked: (swapUpward: boolean, index?: number) => void;
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
  onSwapClicked: (swapUpward: boolean, index?: number) => void;
}

const OrderColumnButtons: React.FC<OrderColumnButtonsProps> = function OrderColumnButtons(
  props: OrderColumnButtonsProps
) {
  const onUpClicked = React.useCallback(() => {
    props.onSwapClicked(true, props.index);
  }, [props]);
  const onDownClicked = React.useCallback(() => {
    props.onSwapClicked(false, props.index);
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

export function useOnSwapClicked<T>(
  state: T[],
  setState: React.Dispatch<React.SetStateAction<T[]>>
): (swapUpward: boolean, index?: number) => void {
  const onSwapClicked = React.useCallback(
    (swapUpward: boolean, index?: number) => {
      if (index == null) {
        return;
      }
      if (swapUpward && index > 0) {
        setState((prev: T[]) => {
          const target = prev[index - 1];
          prev[index - 1] = prev[index];
          prev[index] = target;
          return [...prev];
        });
        return;
      }
      if (!swapUpward && index < state.length - 1) {
        setState((prev: T[]) => {
          const target = prev[index + 1];
          prev[index + 1] = prev[index];
          prev[index] = target;
          return [...prev];
        });
      }
    },
    [state, setState]
  );
  return onSwapClicked;
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
      name: renderToString("order"),
      minWidth: props.orderColumnMinWidth ?? 200,
      maxWidth: props.orderColumnMaxWidth ?? 200,
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
