import React from "react";
import Widget from "./Widget";
import OrderButtons from "./OrderButtons";
import styles from "./WidgetWithOrdering.module.scss";

interface WidgetWithOrderingProps {
  disabled: boolean;
  index: number;
  itemCount: number;
  HeaderMessageComponent?: React.ReactNode;
  HeaderComponent: React.ReactNode;
  onSwapClicked: (index1: number, index2: number) => void;
  renderAriaLabel: (index?: number) => string;
  children: React.ReactNode;
  className?: string;
}

const WidgetWithOrdering: React.FC<WidgetWithOrderingProps> =
  function WidgetWithOrdering(props: WidgetWithOrderingProps) {
    const {
      disabled,
      index,
      itemCount,
      HeaderMessageComponent,
      HeaderComponent,
      onSwapClicked,
      children,
      className,
      renderAriaLabel,
    } = props;

    return (
      <Widget className={className}>
        {HeaderMessageComponent && <div>{HeaderMessageComponent}</div>}
        <div className={styles.header}>
          {HeaderComponent}
          <OrderButtons
            disabled={disabled}
            index={index}
            itemCount={itemCount}
            onSwapClicked={onSwapClicked}
            renderAriaLabel={renderAriaLabel}
          />
        </div>
        {children}
      </Widget>
    );
  };

export default WidgetWithOrdering;
