import { useMemo } from "react";
import { useAppearance } from "./useAppearance";
import { ChartThemeColors, readChartThemeColors } from "../util/themeTokens";

// Returns `read()` and re-runs it when the appearance changes, so canvas
// colors read from CSS variables follow the light/dark mode. `read` must be a
// stable (module-level) function.
export function useThemeColors<T>(read: () => T): T {
  const { resolved } = useAppearance();
  // `resolved` is not read here, but the token values `read` returns depend
  // on the appearance class it reflects.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  return useMemo(() => read(), [read, resolved]);
}

export function useChartThemeColors(): ChartThemeColors {
  return useThemeColors(readChartThemeColors);
}
