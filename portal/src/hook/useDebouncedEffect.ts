import { useEffect } from "react";

export function useDebouncedEffect(effect: () => void, periodMS: number): void {
  useEffect(() => {
    const handle = window.setTimeout(() => {
      effect();
    }, periodMS);
    return () => window.clearTimeout(handle);
  }, [effect, periodMS]);
}
