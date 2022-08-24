import React from "react";
import {
  MessageBar,
  IMessageBarProps,
  useTheme,
  PartialTheme,
  ThemeProvider,
} from "@fluentui/react";

export interface FeatureDisabledMessageBarProps extends IMessageBarProps {}

const FeatureDisabledMessageBar: React.VFC<FeatureDisabledMessageBarProps> =
  function FeatureDisabledMessageBar(props: FeatureDisabledMessageBarProps) {
    const { ...rest } = props;

    const theme = useTheme();
    const newTheme: PartialTheme = {
      semanticColors: {
        messageText: theme.palette.themePrimary,
        messageLink: theme.semanticColors.link,
        messageLinkHovered: theme.semanticColors.linkHovered,
        infoIcon: theme.palette.themePrimary,
      },
    };
    return (
      <ThemeProvider theme={newTheme}>
        <MessageBar
          {...rest}
          styles={{
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
            },
            icon: {
              lineHeight: "20px",
            },
          }}
        >
          {props.children}
        </MessageBar>
      </ThemeProvider>
    );
  };

export default FeatureDisabledMessageBar;
