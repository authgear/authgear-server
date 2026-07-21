import { useCallback, useState } from "react";
import { APIError } from "../error/error";
import { SimpleFormModel } from "./useSimpleForm";
import { useSyncFormStates } from "./useSyncFormStates";

export interface SubmitOutcome<State, Result> {
  result: Result;
  // The new baseline ("initial") state to adopt once this submission
  // succeeds -- pass the state you just saved. Omit if this submission
  // did not change anything worth re-baselining (isDirty is left as-is).
  nextInitialState?: State;
}

export interface UseFormWithExternalInitialStateProps<State, Result> {
  defaultState: State;
  submit: (state: State) => Promise<SubmitOutcome<State, Result>>;
  validate?: (state: State) => APIError | null;
}

// useFormWithExternalInitialState is for forms whose initial state can
// be driven by something outside this hook, such as a GraphQL query
// result that can be refetched independently of this form's own save()
// (e.g. a record that can also be edited elsewhere in the same
// session).
//
// defaultState is synced into initialState/currentState while
// rendering, not in a useEffect: comparing defaultState against a
// tracked previous value and calling setState conditionally during
// render is a React-supported pattern for exactly this case
// ("Adjusting some state when a prop changes" in the React docs), and
// it settles within the same render/commit rather than one render
// later. This matters because save() below can also synchronously
// adopt a new baseline (via submit()'s returned nextInitialState) in
// the very same call as reporting the submission complete -- and both
// of those, together with getIsDirty (see useSyncFormStates), stay
// correct regardless of render timing.
export function useFormWithExternalInitialState<State, Result = unknown>(
  props: UseFormWithExternalInitialStateProps<State, Result>
): SimpleFormModel<State, Result> {
  const { defaultState, submit, validate } = props;

  const {
    initialState,
    currentState,
    setInitialState,
    setCurrentState,
    getIsDirty,
  } = useSyncFormStates<State>(defaultState, defaultState);

  const [prevDefaultState, setPrevDefaultState] = useState(defaultState);
  if (defaultState !== prevDefaultState) {
    setPrevDefaultState(defaultState);
    setInitialState(defaultState);
    setCurrentState(defaultState);
  }

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
      const { result, nextInitialState } = await submit(state);
      setError(null);
      if (nextInitialState !== undefined) {
        setInitialState(nextInitialState);
        setCurrentState(nextInitialState);
      }
      setSubmissionResult(result);
      setIsSubmitted(true);
    } catch (e: unknown) {
      setError(e);
      throw e;
    } finally {
      setIsLoading(false);
    }
  }, [isLoading, submit, validate, state, setInitialState, setCurrentState]);

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
