import { useWindowSize } from "./useWindowSize";

export type ScreenBreakpoint = "mobile" | "tablet" | "desktop";

export function useScreenBreakpoint(): ScreenBreakpoint {
  const { width } = useWindowSize();
  // This should corresponds to tailwind.config.js
  if (width <= 640) {
    return "mobile";
  } else if (width <= 1080) {
    return "tablet";
  }
  return "desktop";
}
