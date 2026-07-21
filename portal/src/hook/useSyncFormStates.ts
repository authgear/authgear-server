import { useCallback, useRef, useState } from "react";
import deepEqual from "deep-equal";

// A single state slot with a synchronous, ref-backed "live" getter.
// Unlike the state value itself (which is only as fresh as the last
// commit, since that's how React re-renders work), the getter always
// reflects the latest value passed to the setter -- immediately,
// before React has had a chance to re-render anything. This makes it
// safe to call from code that isn't gated by React's render/effect
// cycle at all, such as a promise `.then()` continuation.
export function useLiveState<T>(
  defaultValue: T
): [T, (value: T | ((prev: T) => T)) => void, () => T] {
  const [state, setStateState] = useState(defaultValue);
  const ref = useRef(state);

  const setState = useCallback((value: T | ((prev: T) => T)) => {
    // Compute the next value and update the ref synchronously, right
    // here -- not inside the updater callback passed to setStateState,
    // which React does not invoke immediately (it's deferred until
    // React actually processes the update, typically during the next
    // render). Using ref.current as the "prev" source keeps this
    // correct even if setState is called multiple times before React
    // re-renders.
    const next =
      typeof value === "function"
        ? (value as (prev: T) => T)(ref.current)
        : value;
    ref.current = next;
    setStateState(next);
  }, []);

  // getState's identity changes exactly when `state` does (a real,
  // render-triggering update) -- not on every render, and not never.
  // That makes it safe to use as a dependency of a useMemo/useEffect, or
  // as a prop a child component re-renders off of: reference-equality
  // is how React's reactivity works, so a getter whose identity never
  // changes (e.g. useCallback(..., [])) can never signal "the value
  // changed" to anything watching it that way. The function body still
  // always reads the shared ref, so calling *any* version of getState
  // (including one captured in a stale closure) still returns the
  // current value.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const getState = useCallback(() => ref.current, [state]);

  return [state, setState, getState];
}

export interface SyncFormStatesModel<State> {
  initialState: State;
  currentState: State | null;
  setInitialState: (state: State) => void;
  setCurrentState: (
    value: State | null | ((prev: State | null) => State | null)
  ) => void;
  // Always-fresh dirty check, safe to call from anywhere (a promise
  // continuation, a timer, a router's navigation blocker, etc.), unlike
  // a plain `isDirty` boolean computed via `useMemo`, which only
  // reflects reality as of the last render.
  getIsDirty: (isEqual?: (a: State, b: State) => boolean) => boolean;
}

// useSyncFormStates manages the (initialState, currentState) pair for a
// form, where either can be updated as part of a successful save (e.g.
// resetting to a freshly-saved baseline) without waiting for React to
// re-render before that update is observable.
export function useSyncFormStates<State>(
  defaultInitialState: State,
  defaultCurrentState: State | null = null
): SyncFormStatesModel<State> {
  const [initialState, setInitialState, getInitialState] =
    useLiveState(defaultInitialState);
  const [currentState, setCurrentState, getCurrentState] =
    useLiveState<State | null>(defaultCurrentState);

  const getIsDirty = useCallback(
    (
      isEqual: (a: State, b: State) => boolean = (a, b) =>
        deepEqual(a, b, { strict: true })
    ) => {
      const current = getCurrentState();
      if (current == null) {
        return false;
      }
      return !isEqual(getInitialState(), current);
    },
    [getInitialState, getCurrentState]
  );

  return {
    initialState,
    currentState,
    setInitialState,
    setCurrentState,
    getIsDirty,
  };
}
