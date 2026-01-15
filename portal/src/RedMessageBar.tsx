import React, { useMemo } from "react";
import {
  MessageBar,
  IMessageBarProps,
  useTheme,
  MessageBarType,
  PartialTheme,
  ThemeProvider,
} from "@fluentui/react";
import { FormattedMessage } from "./intl";
import { Link as ReactRouterLink, useParams } from "react-router-dom";
import { useMergedStyles } from "./util/mergeStyles";

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

  const styles = useMergedStyles(
    {
      root: {
        selectors: {
          ".ms-Link": {
            // Since both the text and the link are of the same color (errorText),
            // we need to add an underline to the link to make them distinguishable.
            textDecoration: "underline",
          },
        },
      },
    },
    stylesProp
  );

  return (
    <ThemeProvider as={React.Fragment} theme={newTheme}>
      <MessageBar
        messageBarType={MessageBarType.error}
        messageBarIconProps={{
          iconName: "Warning",
        }}
        styles={styles}
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
        values={{
          // eslint-disable-next-line react/no-unstable-nested-components
          ReactRouterLink: (children: React.ReactNode) => (
            <ReactRouterLink to={`/project/${appID}/advanced/sms-gateway`}>
              {children}
            </ReactRouterLink>
          ),
        }}
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

export function RedMessageBar_RemindConfigureSMTPInNonSMTPConfigurationScreen(
  props: IMessageBarProps
): React.ReactElement {
  const { appID } = useParams() as { appID: string };
  return (
    <RedMessageBar {...props}>
      <FormattedMessage
        id="RedMessageBar.remind-configure-smtp-in-non-smtp-configuration-screen"
        values={{
          // eslint-disable-next-line react/no-unstable-nested-components
          ReactRouterLink: (children: React.ReactNode) => (
            <ReactRouterLink to={`/project/${appID}/advanced/smtp`}>
              {children}
            </ReactRouterLink>
          ),
        }}
      />
    </RedMessageBar>
  );
}
