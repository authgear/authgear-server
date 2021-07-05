import React from "react";
import Widget from "./Widget";
import cn from "classnames";
import OrderButtons from "./OrderButtons";
import styles from "./WidgetWithOrdering.module.scss";

interface WidgetWithOrderingProps {
  index: number;
  itemCount: number;
  HeaderMessageComponent?: React.ReactNode;
  HeaderComponent: React.ReactNode;
  onSwapClicked: (index1: number, index2: number) => void;
  renderAriaLabel: (index?: number) => string;
  readOnly?: boolean;
  children: React.ReactNode;
  className?: string;
}

const WidgetWithOrdering: React.FC<WidgetWithOrderingProps> =
  function WidgetWithOrdering(props: WidgetWithOrderingProps) {
    const {
      index,
      itemCount,
      HeaderMessageComponent,
      HeaderComponent,
      onSwapClicked,
      readOnly,
      children,
      className,
      renderAriaLabel,
    } = props;

    return (
      <Widget className={className}>
        {HeaderMessageComponent && <div>{HeaderMessageComponent}</div>}
        <div className={styles.header}>
          {HeaderComponent}
          <div className={cn({ [styles.readOnly]: readOnly })}>
            <OrderButtons
              index={index}
              itemCount={itemCount}
              onSwapClicked={onSwapClicked}
              renderAriaLabel={renderAriaLabel}
            />
          </div>
        </div>
        <div className={cn({ [styles.readOnly]: readOnly })}>{children}</div>
      </Widget>
    );
  };

export default WidgetWithOrdering;
