import React from "react";
import { Text } from "@fluentui/react";

export interface ScreenDescriptionProps {
  className?: string;
  children?: React.ReactNode;
}

const ScreenDescription: React.VFC<ScreenDescriptionProps> =
  function ScreenDescription(props: ScreenDescriptionProps) {
    const { className, children } = props;
    return (
      <Text block={true} className={className}>
        {children}
      </Text>
    );
  };

export default ScreenDescription;
