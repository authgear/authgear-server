import { LocaleProvider } from "@oursky/react-messageformat";
import React from "react";
import ExternalLink from "../../ExternalLink";
import Link from "../../Link";
import { ILinkProps } from "@fluentui/react";
import DEFAULT_MESSAGES from "../../locale-data/en.json";
import { SystemConfig } from "../../system-config";

const DocLink: React.VFC<ILinkProps> = (props: ILinkProps) => {
  return <ExternalLink {...props} />;
};

const defaultComponents = {
  ExternalLink,
  ReactRouterLink: Link,
  DocLink,
  br: () => <br />,
};

export function AppLocaleProvider({
  children,
  systemConfig,
}: {
  systemConfig?: SystemConfig;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <LocaleProvider
      locale="en"
      messageByID={systemConfig?.translations.en ?? DEFAULT_MESSAGES}
      defaultComponents={defaultComponents}
    >
      {children}
    </LocaleProvider>
  );
}
