import React, { useCallback, useMemo } from "react";
import { FormattedMessage as RealFormattedMessage, useIntl } from "react-intl";

// This file is to support legacy API that uses @oursky/react-messageformat

export interface FormattedMessageProps {
  id: string;
  values?: Record<string, any>;
}

export function FormattedMessage(
  props: FormattedMessageProps
): React.ReactElement {
  return <RealFormattedMessage id={props.id} values={props.values} />;
}

export type Values = Record<string, any>;

export interface IntlContextValue {
  locale: string;
  renderToString: (id: string, values?: Record<any, any>) => string;
}

export const Context = React.createContext<IntlContextValue>({
  locale: "en",
  renderToString: (id: string, _values?: Record<any, any>) => id,
});

export function IntlContextProvider({
  children,
}: {
  children: React.ReactNode;
}): React.ReactElement {
  const intl = useIntl();
  const renderToString = useCallback(
    (id: string, values?: Record<any, any>) => {
      return intl.formatMessage({ id }, values);
    },
    [intl]
  );

  return (
    <Context.Provider
      value={useMemo(
        () => ({ locale: intl.locale, renderToString }),
        [intl.locale, renderToString]
      )}
    >
      {children}
    </Context.Provider>
  );
}
