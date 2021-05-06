import { useRef, useEffect, useState } from "react";

export default function useDelayedValue<T>(value: T, delay: number): T {
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>();
  const [state, setState] = useState<T>(value);
  useEffect(() => {
    timeoutRef.current = setTimeout(() => {
      timeoutRef.current = undefined;
      setState(value);
    }, delay);
    return () => {
      if (timeoutRef.current != null) {
        clearTimeout(timeoutRef.current);
        timeoutRef.current = undefined;
      }
    };
  }, [value, delay]);
  return state;
}
