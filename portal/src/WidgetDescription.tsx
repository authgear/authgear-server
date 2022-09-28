import React from "react";
import { Text } from "@fluentui/react";

export interface WidgetDescriptionProps {
  className?: string;
  children?: React.ReactNode;
}

const WidgetDescription: React.VFC<WidgetDescriptionProps> =
  function WidgetDescription(props: WidgetDescriptionProps) {
    const { className, children } = props;
    return (
      <Text
        as="p"
        variant="medium"
        className={className}
        block={true}
        styles={{
          root: {
            // See Widget.
            lineHeight: "20px",
          },
        }}
      >
        {children}
      </Text>
    );
  };

export default WidgetDescription;
