import { Context } from "../intl";
import { useContext, useMemo } from "react";
import ALL_COUNTRIES from "../data/country.json";

export interface Option {
  key: string;
  text: string;
}

export function useMakeAlpha2Options(): {
  alpha2Options: Option[];
} {
  const { renderToString } = useContext(Context);
  const alpha2Options = useMemo(() => {
    const options: { key: string; text: string }[] = ALL_COUNTRIES.map(
      (c: { Alpha2: string }) => {
        const countryName = renderToString(`Territory.${c.Alpha2}`);
        return {
          key: c.Alpha2,
          text: `${c.Alpha2} - ${countryName}`,
        };
      }
    );
    options.sort((o1, o2) => {
      return o1.text.localeCompare(o2.text);
    });
    return options;
  }, [renderToString]);

  return { alpha2Options };
}
