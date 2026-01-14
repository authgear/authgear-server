import React, {
  useEffect,
  useContext,
  useCallback,
  useState,
  useMemo,
  createContext,
} from "react";
import { useId } from "@fluentui/react-hooks";

interface ErrorContextValue {
  errors: Map<string, unknown>;
  error: unknown;
  addError(id: string, error: unknown): void;
  removeError(id: string): void;
}

const DEFAULT_VALUE: ErrorContextValue = {
  errors: new Map(),
  error: null,
  addError: (_id: string, _error: unknown) => { },
  removeError: (_id: string) => { },
};

const ErrorContext = createContext(DEFAULT_VALUE);

export interface ErrorContextProviderProps {
  children?: React.ReactNode;
}

export function ErrorContextProvider(
  props: ErrorContextProviderProps
): React.ReactElement {
  const { children } = props;

  const [state, setState] = useState<
    Omit<ErrorContextValue, "addError" | "removeError">
  >({
    errors: new Map(),
    error: null,
  });

  const addError = useCallback((id: string, incomingError: unknown) => {
    setState((prev) => {
      const map = new Map(prev.errors);
      map.set(id, incomingError);

      let error: unknown = null;
      for (const e of map.values()) {
        if (e != null) {
          error = e;
        }
      }

      return {
        errors: map,
        error,
      };
    });
  }, []);

  const removeError = useCallback((id: string) => {
    setState((prev) => {
      const map = new Map(prev.errors);
      map.delete(id);

      let error: unknown = null;
      for (const e of map.values()) {
        if (e != null) {
          error = e;
        }
      }

      return {
        errors: map,
        error,
      };
    });
  }, []);

  const value = useMemo(() => {
    return {
      ...state,
      addError,
      removeError,
    };
  }, [state, addError, removeError]);

  return (
    <ErrorContext.Provider value={value}>{children}</ErrorContext.Provider>
  );
}

export function useProvideError(error: unknown): void {
  const id = useId();
  // eslint-disable-next-line @typescript-eslint/unbound-method
  const { addError, removeError } = useContext(ErrorContext);
  useEffect(() => {
    addError(id, error);
    return () => {
      removeError(id);
    };
  }, [id, addError, removeError, error]);
}

export function useErrorState(): [
  err: unknown,
  setErr: React.Dispatch<unknown>
] {
  const [err, setErr] = useState<unknown>(undefined);
  useProvideError(err);
  return [err, setErr];
}

export function useConsumeError(): unknown {
  const { error } = useContext(ErrorContext);
  return error;
}
