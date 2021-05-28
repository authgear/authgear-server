import React from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import styles from "./WidgetDescription.module.scss";

export interface WidgetDescriptionProps {
  className?: string;
  children?: React.ReactNode;
}

const WidgetDescription: React.FC<WidgetDescriptionProps> =
  function WidgetDescription(props: WidgetDescriptionProps) {
    const { className, children } = props;
    return (
      <Text
        as="p"
        variant="medium"
        className={cn(className, styles.description)}
      >
        {children}
      </Text>
    );
  };

export default WidgetDescription;
