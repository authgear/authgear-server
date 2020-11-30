import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";

export interface SimpleFormModel<State> {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  isSubmitted: boolean;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reset: () => void;
  save: () => void;
}

export function useSimpleForm<State>(
  defaultState: State,
  submit: (state: State) => Promise<unknown>
): SimpleFormModel<State> {
  const [initialState, setInitialState] = useState(defaultState);
  const [currentState, setCurrentState] = useState(initialState);
  const [error, setError] = useState<unknown>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);

  const isDirty = useMemo(
    () => !deepEqual(initialState, currentState, { strict: true }),
    [initialState, currentState]
  );

  const reset = useCallback(() => {
    if (isLoading || !isDirty) {
      return;
    }
    setError(null);
    setCurrentState(initialState);
  }, [isLoading, isDirty, initialState]);

  const save = useCallback(() => {
    if (isLoading || !isDirty) {
      return;
    }

    setIsLoading(true);
    submit(currentState)
      .then(() => {
        setError(null);
        setInitialState(currentState);
        setIsSubmitted(true);
      })
      .catch((e) => setError(e))
      .finally(() => setIsLoading(false));
  }, [isLoading, isDirty, submit, currentState]);

  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState(fn(currentState));
    },
    [currentState]
  );

  return {
    isUpdating: isLoading,
    isDirty,
    isSubmitted,
    updateError: error,
    state: currentState,
    setState,
    reset,
    save,
  };
}
