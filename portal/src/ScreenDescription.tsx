import React from "react";
import cn from "classnames";
import { DefaultEffects, Text } from "@fluentui/react";
import styles from "./ScreenDescription.module.scss";

export interface ScreenDescriptionProps {
  className?: string;
  children?: React.ReactNode;
}

const ScreenDescription: React.FC<ScreenDescriptionProps> = function ScreenDescription(
  props: ScreenDescriptionProps
) {
  const { className, children } = props;
  return (
    <div
      className={cn(className, styles.description)}
      style={{ boxShadow: DefaultEffects.elevation4 }}
    >
      <Text>{children}</Text>
    </div>
  );
};

export default ScreenDescription;
