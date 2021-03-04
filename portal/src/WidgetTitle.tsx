import React from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import styles from "./WidgetTitle.module.scss";

export interface WidgetTitleProps {
  className?: string;
  children?: React.ReactNode;
}

const WidgetTitle: React.FC<WidgetTitleProps> = function WidgetTitle(
  props: WidgetTitleProps
) {
  const { className, children } = props;
  return (
    <Text as="h2" variant="xLarge" className={cn(className, styles.title)}>
      {children}
    </Text>
  );
};

export default WidgetTitle;
