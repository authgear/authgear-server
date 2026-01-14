import React, { useCallback, useContext } from "react";
import { IconButton, IIconProps } from "@fluentui/react";
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

const DOWN_ICON_PROPS: IIconProps = {
  iconName: "ChevronDown",
  className: styles.icon,
};

const UP_ICON_PROPS: IIconProps = {
  iconName: "ChevronUp",
  className: styles.icon,
};

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
    <div>
      <IconButton
        className={styles.orderButton}
        disabled={disabled || index === itemCount - 1}
        onClick={onDownClicked}
        iconProps={DOWN_ICON_PROPS}
        ariaLabel={renderToString("OrderButtons.move-down")}
      />
      <IconButton
        className={styles.orderButton}
        disabled={disabled || index === 0}
        onClick={onUpClicked}
        iconProps={UP_ICON_PROPS}
        ariaLabel={renderToString("OrderButtons.move-up")}
      />
    </div>
  );
};

export default OrderButtons;
