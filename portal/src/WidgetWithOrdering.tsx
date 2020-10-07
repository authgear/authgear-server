import React from "react";
import { DefaultEffects } from "@fluentui/react";
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
    <div className={className} style={{ boxShadow: DefaultEffects.elevation4 }}>
      <div className={styles.header}>
        <div className={styles.propsHeader}>{HeaderComponent}</div>
        <OrderButtons
          index={index}
          itemCount={itemCount}
          onSwapClicked={onSwapClicked}
          renderAriaLabel={renderAriaLabel}
        />
      </div>
      <div className={styles.contentContainer}>
        <div className={cn(styles.content, { [styles.readOnly]: readOnly })}>
          {children}
        </div>
      </div>
    </div>
  );
};

export default WidgetWithOrdering;
