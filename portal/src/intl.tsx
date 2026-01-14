import React, { useCallback } from 'react'
import { FormattedMessage as RealFormattedMessage, useIntl } from 'react-intl'

// This file is to support legacy API that uses @oursky/react-messageformat

export function FormattedMessage(props: {
  id: string;
  values?: Record<string, any>;
}): React.ReactElement {
  return <RealFormattedMessage {...props} />
}

export type Values = Record<string, any>;

interface IntlContextValue {
  locale: string;
  renderToString: (id: string, values?: Record<any, any>) => string;
}

export const Context = React.createContext<IntlContextValue>({
  locale: "en",
  renderToString: (id: string, _values?: Record<any, any>) => id,
})

export function IntlContextProvider({ children }: { children: React.ReactNode }) {
  const intl = useIntl()
  const renderToString = useCallback((id: string, values?: Record<any, any>) => {
    return intl.formatMessage({ id }, values)
  }, [intl])

  return <Context.Provider value={{ locale: intl.locale, renderToString }}>{children}</Context.Provider>
}