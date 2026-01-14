import { useCallback, useContext } from "react";
import en from "cldr-localenames-full/main/en/territories.json";
import { Context } from "../intl";

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
  if (locale === "en") {
    const territories: Record<string, string | undefined> =
      en.main.en.localeDisplayNames.territories;
    const name = territories[code];
    if (name != null) {
      return name;
    }
  }

  return code;
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
