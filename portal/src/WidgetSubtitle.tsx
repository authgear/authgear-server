import React, { ReactNode, ReactElement } from "react";
import { Heading } from "@radix-ui/themes";

export interface WidgetSubtitleProps {
  children?: ReactNode;
}

const FIELD_TITLE_STYLE: React.CSSProperties = {
  // Match the previous Fluent UI medium variant (semibold).
  fontWeight: 600,
  lineHeight: "20px",
};

export default function WidgetSubtitle(
  props: WidgetSubtitleProps
): ReactElement {
  const { children } = props;
  return (
    <Heading as="h3" size="2" style={FIELD_TITLE_STYLE}>
      {children}
    </Heading>
  );
}
