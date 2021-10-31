import React from "react";
import { Text } from "@fluentui/react";

export interface WidgetDescriptionProps {
  className?: string;
  children?: React.ReactNode;
}

const WidgetDescription: React.FC<WidgetDescriptionProps> =
  function WidgetDescription(props: WidgetDescriptionProps) {
    const { className, children } = props;
    return (
      <Text as="p" variant="medium" className={className} block={true}>
        {children}
      </Text>
    );
  };

export default WidgetDescription;
