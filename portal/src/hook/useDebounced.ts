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

    // eslint-disable-next-line react-hooks/set-state-in-effect
    setIsDebouncing(true);
    const handle = setTimeout(() => {
      setDebouncedValue(value);
      setIsDebouncing(false);
    }, periodMS);
    return () => clearTimeout(handle);
  }, [debouncedValue, value, periodMS]);

  return [debouncedValue, isDebouncing];
}
