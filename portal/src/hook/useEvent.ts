import React from "react";

type AnyFunction = (...args: any[]) => any;

// Basically an implementation of useEvent in React 18
// See https://www.reactuse.com/effect/useevent/
// The implementation copied from https://github.com/scottrippey/react-use-event-hook/blob/main/src/useEvent.ts
// with some modifications
export function useEvent<TCallback extends AnyFunction>(
  callback: TCallback
): TCallback {
  // Keep track of the latest callback:
  const latestRef = React.useRef<TCallback>(
    useEvent_shouldNotBeInvokedBeforeMount as any
  );
  latestRef.current = callback;

  // Create a stable callback that always calls the latest callback:
  // using useRef instead of useCallback avoids creating and empty array on every render
  const stableRef = React.useRef<TCallback>(null as any);
  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
  if (!stableRef.current) {
    stableRef.current = function (this: any) {
      // eslint-disable-next-line prefer-rest-params
      return latestRef.current.apply(this, arguments as any);
    } as TCallback;
  }

  return stableRef.current;
}

/**
 * Render methods should be pure, especially when concurrency is used,
 * so we will throw this error if the callback is called while rendering.
 */
function useEvent_shouldNotBeInvokedBeforeMount() {
  throw new Error(
    "INVALID_USEEVENT_INVOCATION: the callback from useEvent cannot be invoked before the component has mounted."
  );
}
