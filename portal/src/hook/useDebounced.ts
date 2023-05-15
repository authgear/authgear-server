import { useEffect, useState } from "react";

export function useDebounced<T>(
  value: T,
  periodMS: number
): [debounced: T, isDebouncing: boolean] {
  const [debouncedValue, setDebouncedValue] = useState(value);
  const [isDebouncing, setIsDebouncing] = useState(false);

  useEffect(() => {
    if (debouncedValue === value) {
      return () => {};
    }

    setIsDebouncing(true);
    const handle = setTimeout(() => {
      setDebouncedValue(value);
      setIsDebouncing(false);
    }, periodMS);
    return () => clearTimeout(handle);
  }, [debouncedValue, value, periodMS]);

  return [debouncedValue, isDebouncing];
}
