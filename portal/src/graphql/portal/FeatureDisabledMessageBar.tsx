import React, { useMemo } from "react";
import {
  MessageBar,
  IMessageBarProps,
  useTheme,
  PartialTheme,
  ThemeProvider,
} from "@fluentui/react";
import { FormattedMessage, Values } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";

export interface FeatureDisabledMessageBarProps extends IMessageBarProps {
  messageID: string;
  messageValues?: Values;
}

const FeatureDisabledMessageBar: React.VFC<FeatureDisabledMessageBarProps> =
  function FeatureDisabledMessageBar(props: FeatureDisabledMessageBarProps) {
    const { messageID, messageValues, ...rest } = props;
    const { appID } = useParams() as { appID: string };

    const theme = useTheme();
    const newTheme: PartialTheme = {
      semanticColors: {
        messageText: theme.palette.themePrimary,
        messageLink: theme.semanticColors.link,
        messageLinkHovered: theme.semanticColors.linkHovered,
        infoIcon: theme.palette.themePrimary,
      },
    };

    const values = useMemo(() => {
      return {
        planPagePath: `/project/${appID}/billing`,
        ...messageValues,
      };
    }, [appID, messageValues]);

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
          <FormattedMessage id={messageID} values={values} />
        </MessageBar>
      </ThemeProvider>
    );
  };

export default FeatureDisabledMessageBar;
