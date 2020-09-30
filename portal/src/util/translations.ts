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

export function useGetTelecomCountryName(): {
  getTelecomCountryName: (code: string) => string;
} {
  const { locale, renderToString } = useContext(Context);
  const getTelecomCountryName = useCallback(
    (code: string) => {
      if (code === "INTERNATIONAL") {
        return renderToString("calling-code-area.international");
      }
      if (code === "GMSS") {
        return renderToString("calling-code-area.gmss");
      }
      return getCountryName(code, locale);
    },
    [locale, renderToString]
  );

  return { getTelecomCountryName };
}
