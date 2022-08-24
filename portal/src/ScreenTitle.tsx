import React from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import styles from "./ScreenTitle.module.css";

export interface ScreenTitleProps {
  className?: string;
  children?: React.ReactNode;
}

const ScreenTitle: React.VFC<ScreenTitleProps> = function ScreenTitle(
  props: ScreenTitleProps
) {
  const { className, children } = props;
  return (
    <Text
      as="h1"
      variant="xxLarge"
      block={true}
      className={cn(styles.title, className)}
    >
      {children}
    </Text>
  );
};

export default ScreenTitle;
