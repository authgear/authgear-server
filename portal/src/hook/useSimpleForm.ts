import { useCallback, useMemo, useState, useEffect } from "react";
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
  save: () => Promise<void>;
}

export interface UseSimpleFormProps<State, Result> {
  defaultState: State;
  submit: (state: State) => Promise<Result>;
  stateMode: // This state mode is for forms where multiple submission is desired.
  // For each submission, the form should be reset to initial state.
  | "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave"
    // This state mode is for forms where the state will be updated externally after save.
    // For example, the updated_at of user profile form will be updated after save.
    // So the initial state must be able to be re-initialized again.
    | "UpdateInitialStateWithUseEffect";
  validate?: (state: State) => APIError | null;
}

export function useSimpleForm<State, Result = unknown>(
  props: UseSimpleFormProps<State, Result>
): SimpleFormModel<State, Result> {
  const { defaultState, stateMode, submit, validate } = props;

  const [initialState, setInitialState] = useState(defaultState);
  const [currentState, setCurrentState] = useState(initialState);
  const [error, setError] = useState<unknown>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [submissionResult, setSubmissionResult] = useState<unknown>(null);

  useEffect(() => {
    if (stateMode === "UpdateInitialStateWithUseEffect") {
      setInitialState(defaultState);
      setCurrentState(defaultState);
    }
  }, [stateMode, defaultState]);

  const isDirty = useMemo(
    () => !deepEqual(initialState, currentState, { strict: true }),
    [initialState, currentState]
  );

  const reset = useCallback(() => {
    if (isLoading) {
      return;
    }
    setError(null);
    setCurrentState(initialState);
  }, [isLoading, initialState]);

  const save = useCallback(async () => {
    if (isLoading) {
      return;
    }

    const err = validate?.(currentState);
    if (err) {
      setError(err);
      // eslint-disable-next-line @typescript-eslint/no-throw-literal
      throw err;
    }

    setIsLoading(true);
    try {
      const result = await submit(currentState);
      setError(null);
      if (
        stateMode ===
        "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave"
      ) {
        setCurrentState(initialState);
      }
      setSubmissionResult(result);
      setIsSubmitted(true);
    } catch (e: unknown) {
      setError(e);
      throw e;
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, submit, validate, currentState, initialState, stateMode]);

  const setState = useCallback((fn: (state: State) => State) => {
    setCurrentState((s) => fn(s));
  }, []);

  return {
    isUpdating: isLoading,
    isDirty,
    isSubmitted,
    submissionResult: submissionResult as Result,
    updateError: error,
    state: currentState,
    setState,
    reset,
    save,
  };
}
