import { useCallback, useRef, useEffect } from "react";

interface NextRenderPromise {
  promise: Promise<void>;
  resolve: () => void;
}

// useAsyncSetState wraps a setState function,
// so that it returns a promise which resolves in the next render.
// This can be used to ensure the latest state is populated before performing the next step.
// An example is:
// await form.setStateAsync(...);
// form.save();
export function useAsyncSetState<S>(
  syncSetState: (fn: (prevState: S) => S) => void
): (fn: (prevState: S) => S) => Promise<void> {
  const promiseRef = useRef<NextRenderPromise | null>(null);

  const asyncSetState = useCallback(
    async (fn: (prevState: S) => S) => {
      if (promiseRef.current == null) {
        let resolve: (() => void) | undefined;
        const newPromise = new Promise<void>((r) => {
          resolve = r;
        });
        if (resolve == null) {
          // Should not happen, assert non null.
          throw Error("unexpected promise resolver is null");
        }
        promiseRef.current = {
          promise: newPromise,
          resolve: resolve,
        };
      }
      syncSetState(fn);
      return promiseRef.current.promise;
    },
    [syncSetState]
  );

  useEffect(() => {
    if (promiseRef.current == null) {
      return;
    }
    promiseRef.current.resolve();
    promiseRef.current = null;
  });

  return asyncSetState;
}
