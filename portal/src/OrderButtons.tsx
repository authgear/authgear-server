import React, { useCallback, useContext } from "react";
import { IconButton } from "@radix-ui/themes";
import { ChevronDownIcon, ChevronUpIcon } from "@radix-ui/react-icons";
import { Context } from "./intl";

import styles from "./OrderButtons.module.css";

interface OrderButtonsProps {
  index?: number;
  disabled: boolean;
  itemCount: number;
  onSwapClicked: (index1: number, index2: number) => void;
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

const OrderButtons: React.VFC<OrderButtonsProps> = function OrderButtons(
  props: OrderButtonsProps
) {
  const { index, disabled, itemCount, onSwapClicked } = props;
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

  return (
    <div className={styles.orderButtons}>
      <IconButton
        type="button"
        className={styles.orderButton}
        variant="ghost"
        color="gray"
        size="1"
        disabled={disabled || index === itemCount - 1}
        onClick={onDownClicked}
        aria-label={renderToString("OrderButtons.move-down")}
      >
        <ChevronDownIcon width="1rem" height="1rem" />
      </IconButton>
      <IconButton
        type="button"
        className={styles.orderButton}
        variant="ghost"
        color="gray"
        size="1"
        disabled={disabled || index === 0}
        onClick={onUpClicked}
        aria-label={renderToString("OrderButtons.move-up")}
      >
        <ChevronUpIcon width="1rem" height="1rem" />
      </IconButton>
    </div>
  );
};

export default OrderButtons;
