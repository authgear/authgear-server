import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";

// useLocationEffect pops the state out from location and perform effect.
// This hook ensures that refreshing a screen will NOT perform the effect again.
export function useLocationEffect<S>(
  effectFunction: (state: S) => void
): S | undefined | null {
  const { state } = useLocation();
  const navigate = useNavigate();
  useEffect(() => {
    if (state != null) {
      effectFunction(state as any);
      navigate("", {
        replace: true,
      });
    }
  }, [effectFunction, navigate, state]);
  return state as any;
}
