import React from "react";
import { MessageBar, IMessageBarProps, useTheme } from "@fluentui/react";

export interface FeatureDisabledMessageBarProps extends IMessageBarProps {}

const FeatureDisabledMessageBar: React.FC<FeatureDisabledMessageBarProps> =
  function FeatureDisabledMessageBar(props: FeatureDisabledMessageBarProps) {
    const { ...rest } = props;

    const theme = useTheme();
    return (
      <MessageBar
        {...rest}
        styles={{
          root: {
            background: theme.palette.themeLighter,
          },
          innerText: {
            fontSize: "14px",
            lineHeight: "20px",
            color: theme.palette.themePrimary,
            a: {
              fontWeight: 600,
              whiteSpace: "nowrap",
              color: `${theme.palette.themePrimary} !important`,
            },
          },
          icon: {
            lineHeight: "20px",
            color: theme.palette.themePrimary,
          },
        }}
      >
        {props.children}
      </MessageBar>
    );
  };

export default FeatureDisabledMessageBar;
