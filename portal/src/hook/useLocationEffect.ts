import { useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";

// useLocationEffect pops the state out from location and perform effect.
// This hook ensures that refreshing a screen will NOT perform the effect again.
export function useLocationEffect<S>(
  effectFunction: (state: S) => void
): S | undefined | null {
  const location = useLocation();
  const { state } = location;
  const navigate = useNavigate();
  useEffect(() => {
    if (state != null) {
      effectFunction(state);
      navigate(location, {
        replace: true,
      });
    }
  }, [effectFunction, navigate, location, state]);
  return state;
}
