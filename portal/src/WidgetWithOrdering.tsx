import React from "react";
import Widget from "./Widget";
import cn from "classnames";
import OrderButtons from "./OrderButtons";
import styles from "./WidgetWithOrdering.module.scss";

interface WidgetWithOrderingProps {
  index: number;
  itemCount: number;
  HeaderComponent: React.ReactNode;
  onSwapClicked: (index1: number, index2: number) => void;
  renderAriaLabel: (index?: number) => string;
  readOnly?: boolean;
  children: React.ReactNode;
  className?: string;
}

const WidgetWithOrdering: React.FC<WidgetWithOrderingProps> = function WidgetWithOrdering(
  props: WidgetWithOrderingProps
) {
  const {
    index,
    itemCount,
    HeaderComponent,
    onSwapClicked,
    readOnly,
    children,
    className,
    renderAriaLabel,
  } = props;

  return (
    <Widget className={className}>
      <div className={styles.header}>
        {HeaderComponent}
        <OrderButtons
          index={index}
          itemCount={itemCount}
          onSwapClicked={onSwapClicked}
          renderAriaLabel={renderAriaLabel}
        />
      </div>
      <div className={cn({ [styles.readOnly]: readOnly })}>{children}</div>
    </Widget>
  );
};

export default WidgetWithOrdering;
