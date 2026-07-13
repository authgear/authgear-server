import { useCallback, useMemo, useState } from "react";
import deepEqual from "deep-equal";
import { APIError } from "../error/error";

export interface SimpleFormModel<State, Result = unknown> {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  isSubmitted: boolean;
  submissionResult: Result | undefined;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reset: () => void;
  save: () => Promise<void>;
}

export interface UseSimpleFormProps<State, Result> {
  defaultState: State;
  submit: (state: State) => Promise<Result>;
  validate?: (state: State) => APIError | null;
}

// useSimpleForm is for forms where the initial state is constant for
// the lifetime of the component and multiple submissions are desired
// (e.g. a "create X" form): after each successful save, currentState
// resets back to that same constant initialState.
//
// For forms whose initial state can also be driven by something
// outside this hook (e.g. a GraphQL query result that can be refetched
// independently of this form's own save()), use
// useFormWithExternalInitialState instead.
export function useSimpleForm<State, Result = unknown>(
  props: UseSimpleFormProps<State, Result>
): SimpleFormModel<State, Result> {
  const { defaultState, submit, validate } = props;

  const [initialState] = useState(defaultState);
  const [currentState, setCurrentState] = useState(initialState);
  const [error, setError] = useState<unknown>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [submissionResult, setSubmissionResult] = useState<unknown>(null);

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
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw err;
    }

    setIsLoading(true);
    try {
      const result = await submit(currentState);
      setError(null);
      setCurrentState(initialState);
      // Since react 18, state updates could be batched,
      // causing bugs like NavigationBlockerDialog showing because isDirty is not updated.
      // Therefore we wait for next tick to ensure latest states are available
      setTimeout(() => {
        setSubmissionResult(result);
        setIsSubmitted(true);
      });
    } catch (e: unknown) {
      setError(e);
      throw e;
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, submit, validate, currentState, initialState]);

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
