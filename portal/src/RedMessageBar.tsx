import React from "react";
import { FormattedMessage } from "./intl";
import { Link as ReactRouterLink, useParams } from "react-router-dom";
import { Callout } from "./components/v2/Callout/Callout";

export interface RedMessageBarProps {
  className?: string;
  children?: React.ReactNode;
}

export default function RedMessageBar(
  props: RedMessageBarProps
): React.ReactElement {
  const { className, children } = props;
  return (
    <Callout
      className={className}
      type="error"
      showCloseButton={false}
      text={children}
    />
  );
}

export function RedMessageBar_RemindConfigureSMSProviderInNonSMSProviderScreen(
  props: RedMessageBarProps
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
  props: RedMessageBarProps
): React.ReactElement {
  return (
    <RedMessageBar {...props}>
      <FormattedMessage id="RedMessageBar.remind-configure-sms-provider-in-sms-provider-screen" />
    </RedMessageBar>
  );
}

export function RedMessageBar_RemindConfigureSMTPInSMTPConfigurationScreen(
  props: RedMessageBarProps
): React.ReactElement {
  return (
    <RedMessageBar {...props}>
      <FormattedMessage id="RedMessageBar.remind-configure-smtp-in-smtp-configuration-screen" />
    </RedMessageBar>
  );
}

export function RedMessageBar_RemindConfigureSMTPInNonSMTPConfigurationScreen(
  props: RedMessageBarProps
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
