import React, { useMemo } from "react";
import { Text } from "@radix-ui/themes";

export interface WidgetDescriptionProps {
  className?: string;
  children?: React.ReactNode;
  styles?: { root?: React.CSSProperties };
}

const DEFAULT_STYLE: React.CSSProperties = {
  // See Widget.
  lineHeight: "20px",
};

const WidgetDescription: React.VFC<WidgetDescriptionProps> =
  function WidgetDescription(props: WidgetDescriptionProps) {
    const { className, children, styles } = props;
    const style = useMemo(
      () => ({ ...DEFAULT_STYLE, ...styles?.root }),
      [styles]
    );
    return (
      <Text as="p" size="2" className={className} style={style}>
        {children}
      </Text>
    );
  };

export default WidgetDescription;
