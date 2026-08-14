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
      // navigate() is async, so we must await it before running
      // effectFunction. Otherwise, if effectFunction triggers its own
      // re-render (e.g. via a form's setState) before navigate() has
      // actually cleared location.state, that render still observes the
      // old, non-null state, and this effect runs again.
      Promise.resolve(
        navigate(location, {
          replace: true,
        })
      )
        .then(() => {
          effectFunction(state);
        })
        .catch((e: unknown) => {
          console.error(e);
        });
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  return state;
}
