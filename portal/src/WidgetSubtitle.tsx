import React, { ReactNode, ReactElement } from "react";
import { Text } from "@fluentui/react";

export interface WidgetSubtitleProps {
  children?: ReactNode;
}

const FIELD_TITLE_STYLES = {
  root: {
    fontWeight: "600",
  },
};

export default function WidgetSubtitle(
  props: WidgetSubtitleProps
): ReactElement {
  const { children } = props;
  return (
    <Text as="h3" block={true} variant="medium" styles={FIELD_TITLE_STYLES}>
      {children}
    </Text>
  );
}
