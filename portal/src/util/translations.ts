import { useCallback, useContext } from "react";
import { getName } from "i18n-iso-countries";
import { Context } from "@oursky/react-messageformat";

export function useGetCountryName(): {
  getCountryName: (code: string) => string;
} {
  const { locale } = useContext(Context);
  const _getCountryName = useCallback(
    (code: string) => {
      return getCountryName(code, locale);
    },
    [locale]
  );

  return {
    getCountryName: _getCountryName,
  };
}

export function getCountryName(code: string, locale: string): string {
  // override library output here
  return getName(code, locale);
}
