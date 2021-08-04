import React from "react";
import { Text } from "@fluentui/react";
import styles from "./WidgetTitle.module.scss";

export interface WidgetTitleProps {
  className?: string;
  children?: React.ReactNode;
  id?: string;
}

const WidgetTitle: React.FC<WidgetTitleProps> = function WidgetTitle(
  props: WidgetTitleProps
) {
  const { className, children, id } = props;
  const element = (
    <Text as="h2" variant="xLarge" className={styles.title}>
      {children}
    </Text>
  );

  if (id != null) {
    return (
      <a id={id} href={"#" + id} className={className}>
        {element}
      </a>
    );
  }

  return <div className={className}>{element}</div>;
};

export default WidgetTitle;
