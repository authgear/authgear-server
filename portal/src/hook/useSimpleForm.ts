import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import { APIError } from "../error/error";

export interface SimpleFormModel<State, Result = unknown> {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  isSubmitted: boolean;
  submissionResult: Result;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reset: () => void;
  save: () => void;
}

export function useSimpleForm<State, Result = unknown>(
  defaultState: State,
  submit: (state: State) => Promise<Result>,
  validate?: (state: State) => APIError | null
): SimpleFormModel<State, Result> {
  const [initialState, setInitialState] = useState(defaultState);
  const [currentState, setCurrentState] = useState(initialState);
  const [error, setError] = useState<unknown>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [submissionResult, setSubmissionResult] = useState<unknown>(null);

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

    const err = validate?.(currentState);
    if (err) {
      setError(err);
      return;
    }

    setIsLoading(true);
    submit(currentState)
      .then((result) => {
        setError(null);
        setInitialState(currentState);
        setSubmissionResult(result);
      })
      .catch((e) => setError(e))
      .finally(() => setIsLoading(false));
  }, [isLoading, isDirty, submit, validate, currentState]);

  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState(fn(currentState));
    },
    [currentState]
  );

  return {
    isUpdating: isLoading,
    isDirty,
    isSubmitted: submissionResult != null,
    submissionResult: submissionResult as Result,
    updateError: error,
    state: currentState,
    setState,
    reset,
    save,
  };
}
