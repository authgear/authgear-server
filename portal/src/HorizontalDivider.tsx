import React, { ReactElement } from "react";
import cn from "classnames";
import { useTheme } from "@fluentui/react";
import styles from "./HorizontalDivider.module.css";

export interface HorizontalDividerProps {
  className?: string;
}

// There is VerticalDivider from Fluent UI, but no HorizontalDivider :(
// So we build one ourselves.
export default function HorizontalDivider(
  props: HorizontalDividerProps
): ReactElement {
  const { className } = props;
  const theme = useTheme();
  return (
    <hr
      className={cn(styles.root, className)}
      style={{
        borderTopColor: theme.palette.neutralLight,
        backgroundColor: theme.palette.neutralLight,
      }}
    />
  );
}
