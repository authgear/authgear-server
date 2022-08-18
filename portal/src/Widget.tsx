import React from "react";
import cn from "classnames";
import { DefaultEffects } from "@fluentui/react";
import styles from "./Widget.module.css";

export interface WidgetProps {
  className?: string;
  children?: React.ReactNode;
}

const Widget: React.FC<WidgetProps> = function Widget(props: WidgetProps) {
  const { className, children } = props;
  return (
    <div
      className={cn(className, styles.root, "mobile:col-span-full")}
      style={{
        boxShadow: DefaultEffects.elevation4,
      }}
    >
      {children}
    </div>
  );
};

export default Widget;
