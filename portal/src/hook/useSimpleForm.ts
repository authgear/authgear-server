import { useCallback, useState } from "react";
import { APIError } from "../error/error";
import { useSyncFormStates } from "./useSyncFormStates";

export interface SimpleFormModel<State, Result = unknown> {
  updateError: unknown;
  isUpdating: boolean;
  isSubmitted: boolean;
  submissionResult: Result | undefined;
  state: State;
  setState: (fn: (state: State) => State) => void;
  reset: () => void;
  save: () => Promise<void>;
  // Always-fresh dirty check, safe to call from anywhere (see
  // useSyncFormStates). For a render-time boolean (e.g. disabling a
  // button), derive one locally: useMemo(() => getIsDirty(), [getIsDirty]).
  getIsDirty: () => boolean;
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

  const { initialState, currentState, setCurrentState, getIsDirty } =
    useSyncFormStates<State>(defaultState, defaultState);
  const [error, setError] = useState<unknown>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [submissionResult, setSubmissionResult] = useState<unknown>(null);

  const state = currentState ?? initialState;

  const reset = useCallback(() => {
    if (isLoading) {
      return;
    }
    setError(null);
    setCurrentState(initialState);
  }, [isLoading, initialState, setCurrentState]);

  const save = useCallback(async () => {
    if (isLoading) {
      return;
    }

    const err = validate?.(state);
    if (err) {
      setError(err);
      // eslint-disable-next-line @typescript-eslint/only-throw-error
      throw err;
    }

    setIsLoading(true);
    try {
      const result = await submit(state);
      setError(null);
      setCurrentState(initialState);
      setSubmissionResult(result);
      setIsSubmitted(true);
    } catch (e: unknown) {
      setError(e);
      throw e;
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, submit, validate, state, initialState, setCurrentState]);

  const setState = useCallback(
    (fn: (state: State) => State) => {
      setCurrentState((s) => fn(s ?? initialState));
    },
    [setCurrentState, initialState]
  );

  return {
    isUpdating: isLoading,
    isSubmitted,
    submissionResult: submissionResult as Result,
    updateError: error,
    state,
    setState,
    reset,
    save,
    getIsDirty,
  };
}
