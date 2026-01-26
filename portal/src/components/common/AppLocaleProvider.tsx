import { IntlProvider } from "react-intl";
import React from "react";
import DEFAULT_MESSAGES from "../../locale-data/en.json";
import { SystemConfig } from "../../system-config";
import { IntlContextProvider } from "../../intl";

const defaultRichTextElements = {
  br: () => <br />,
  b: (children: React.ReactNode) => <b>{children}</b>,
  strong: (children: React.ReactNode) => <strong>{children}</strong>,
  code: (children: React.ReactNode) => <code>{children}</code>,
  pre: (children: React.ReactNode) => <pre>{children}</pre>,
  small: (children: React.ReactNode) => <small>{children}</small>,
};

// This is to support legacy API that uses @oursky/react-messageformat
function _LegacyAPIContextProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  return <IntlContextProvider>{children}</IntlContextProvider>;
}

export function AppLocaleProvider({
  children,
  systemConfig,
}: {
  systemConfig?: SystemConfig;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <IntlProvider
      locale="en"
      messages={systemConfig?.translations.en ?? DEFAULT_MESSAGES}
      defaultRichTextElements={defaultRichTextElements}
    >
      <_LegacyAPIContextProvider>{children}</_LegacyAPIContextProvider>
    </IntlProvider>
  );
}
