import React, { useMemo } from "react";
import {
  MessageBar,
  IMessageBarProps,
  useTheme,
  PartialTheme,
  ThemeProvider,
} from "@fluentui/react";
import { useMergedStyles } from "./util/mergeStyles";

export default function BlueMessageBar(
  props: IMessageBarProps
): React.ReactElement {
  const theme = useTheme();
  const newTheme: PartialTheme = useMemo(
    () => ({
      semanticColors: {
        messageText: theme.palette.themePrimary,
        messageLink: theme.semanticColors.link,
        messageLinkHovered: theme.semanticColors.linkHovered,
        infoIcon: theme.palette.themePrimary,
      },
    }),
    [
      theme.palette.themePrimary,
      theme.semanticColors.link,
      theme.semanticColors.linkHovered,
    ]
  );

  const { styles: stylesProp, ...rest } = props;

  const styles = useMergedStyles(
    {
      root: {
        background: theme.palette.themeLighter,
      },
      innerText: {
        fontSize: "14px",
        lineHeight: "20px",
        a: {
          fontWeight: 600,
          whiteSpace: "nowrap",
        },
        span: {
          "a:last-child": {
            padding: 0,
          },
        },
      },
      icon: {
        lineHeight: "20px",
      },
    },
    stylesProp
  );

  return (
    <ThemeProvider as={React.Fragment} theme={newTheme}>
      <MessageBar styles={styles} {...rest} />
    </ThemeProvider>
  );
}
