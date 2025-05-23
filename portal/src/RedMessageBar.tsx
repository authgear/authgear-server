import React, { useMemo } from "react";
import {
  MessageBar,
  IMessageBarProps,
  useTheme,
  MessageBarType,
  PartialTheme,
  ThemeProvider,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";

export default function RedMessageBar(
  props: IMessageBarProps
): React.ReactElement {
  const theme = useTheme();
  const newTheme: PartialTheme = useMemo(
    () => ({
      semanticColors: {
        messageText: theme.semanticColors.errorText,
        messageLink: theme.semanticColors.errorText,
        messageLinkHovered: theme.semanticColors.errorText,
      },
    }),
    [theme.semanticColors.errorText]
  );
  const { styles: stylesProp, ...rest } = props;
  return (
    <ThemeProvider as={React.Fragment} theme={newTheme}>
      <MessageBar
        messageBarType={MessageBarType.error}
        messageBarIconProps={{
          iconName: "Warning",
        }}
        styles={stylesProp}
        {...rest}
      />
    </ThemeProvider>
  );
}

export function RedMessageBar_RemindConfigureSMSProviderInNonSMSProviderScreen(
  props: IMessageBarProps
): React.ReactElement {
  const { appID } = useParams() as { appID: string };
  return (
    <RedMessageBar {...props}>
      <FormattedMessage
        id="RedMessageBar.remind-configure-sms-provider-in-non-sms-provider-screen"
        values={{ to: `/project/${appID}/advanced/sms-gateway` }}
      />
    </RedMessageBar>
  );
}

export function RedMessageBar_RemindConfigureSMSProviderInSMSProviderScreen(
  props: IMessageBarProps
): React.ReactElement {
  return (
    <RedMessageBar {...props}>
      <FormattedMessage id="RedMessageBar.remind-configure-sms-provider-in-sms-provider-screen" />
    </RedMessageBar>
  );
}

export function RedMessageBar_RemindConfigureSMTPInSMTPConfigurationScreen(
  props: IMessageBarProps
): React.ReactElement {
  return (
    <RedMessageBar {...props}>
      <FormattedMessage id="RedMessageBar.remind-configure-smtp-in-smtp-configuration-screen" />
    </RedMessageBar>
  );
}
