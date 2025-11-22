import React from "react";
import { Text, ITextProps } from "@fluentui/react";
import { useMergedStylesPlain } from "./util/mergeStyles";

export interface WidgetDescriptionProps {
  className?: string;
  children?: React.ReactNode;
  styles?: ITextProps["styles"];
}

const DEFAULT_STYLES: ITextProps["styles"] = {
  root: {
    // See Widget.
    lineHeight: "20px",
  },
};

const WidgetDescription: React.VFC<WidgetDescriptionProps> =
  function WidgetDescription(props: WidgetDescriptionProps) {
    const { className, children, styles } = props;
    const mergedStyles = useMergedStylesPlain(DEFAULT_STYLES, styles);
    return (
      <Text
        as="p"
        variant="medium"
        className={className}
        block={true}
        // @ts-expect-error
        styles={mergedStyles}
      >
        {children}
      </Text>
    );
  };

export default WidgetDescription;
