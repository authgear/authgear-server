import React, { useCallback, useContext, useMemo } from "react";
import { IconButton } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";

import styles from "./OrderButtons.module.scss";

interface OrderButtonsProps {
  index?: number;
  itemCount: number;
  onSwapClicked: (index1: number, index2: number) => void;
  renderAriaLabel: (index?: number) => string;
}

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

const OrderButtons: React.FC<OrderButtonsProps> = function OrderButtons(
  props: OrderButtonsProps
) {
  const { index, itemCount, onSwapClicked, renderAriaLabel } = props;
  const { renderToString } = useContext(Context);
  const onUpClicked = useCallback(() => {
    if (index == null) {
      return;
    }
    onSwapClicked(index, index - 1);
  }, [index, onSwapClicked]);
  const onDownClicked = useCallback(() => {
    if (index == null) {
      return;
    }
    onSwapClicked(index, index + 1);
  }, [index, onSwapClicked]);

  const ariaLabelUp = useMemo(() => {
    return renderToString("OrderButtons.move-up", {
      key: renderAriaLabel(index),
    });
  }, [renderAriaLabel, index, renderToString]);
  const ariaLabelDown = useMemo(() => {
    return renderToString("OrderButtons.move-down", {
      key: renderAriaLabel(index),
    });
  }, [renderAriaLabel, index, renderToString]);

  return (
    <div>
      <IconButton
        className={styles.orderButton}
        disabled={index === itemCount - 1}
        onClick={onDownClicked}
        iconProps={{ iconName: "ChevronDown" }}
        ariaLabel={ariaLabelDown}
      />
      <IconButton
        className={styles.orderButton}
        disabled={index === 0}
        onClick={onUpClicked}
        iconProps={{ iconName: "ChevronUp" }}
        ariaLabel={ariaLabelUp}
      />
    </div>
  );
};

export default OrderButtons;
