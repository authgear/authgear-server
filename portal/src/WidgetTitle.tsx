import React from "react";
import { Text } from "@fluentui/react";

export interface WidgetTitleProps {
  className?: string;
  children?: React.ReactNode;
  id?: string;
}

const WidgetTitle: React.VFC<WidgetTitleProps> = function WidgetTitle(
  props: WidgetTitleProps
) {
  const { className, children, id } = props;
  const element = (
    <Text
      as="h2"
      variant="xLarge"
      block={true}
      styles={{
        root: {
          // 16px padding + 28px + 20px gap = 64px
          // Widget's collapsed height is related to this value.
          lineHeight: "28px",
        },
      }}
    >
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
