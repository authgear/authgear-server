import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useMemo,
  useEffect,
} from "react";
import { useId } from "@fluentui/react-hooks";

interface LoadingContextValue {
  loadables: Map<string, boolean>;
  isLoading: boolean;
  addLoadable(id: string, isLoading: boolean): void;
  removeLoadable(id: string): void;
}

const DEFAULT_VALUE: LoadingContextValue = {
  loadables: new Map(),
  isLoading: false,
  addLoadable: (_id: string, _isLoading: boolean) => {},
  removeLoadable: (_id: string) => {},
};

const LoadingContext = createContext(DEFAULT_VALUE);

export interface LoadingContextProviderProps {
  children?: React.ReactNode;
}

export function LoadingContextProvider(
  props: LoadingContextProviderProps
): React.ReactElement {
  const { children } = props;

  const [state, setState] = useState<
    Omit<LoadingContextValue, "addLoadable" | "removeLoadable">
  >({
    loadables: new Map(),
    isLoading: false,
  });

  const addLoadable = useCallback((id: string, isLoading: boolean) => {
    setState((prev) => {
      const map = new Map(prev.loadables);
      map.set(id, isLoading);

      let loading = false;
      for (const i of map.values()) {
        if (i) {
          loading = true;
        }
      }

      return {
        loadables: map,
        isLoading: loading,
      };
    });
  }, []);

  const removeLoadable = useCallback((id: string) => {
    setState((prev) => {
      const map = new Map(prev.loadables);
      map.delete(id);

      let loading = false;
      for (const i of map.values()) {
        if (i) {
          loading = true;
        }
      }

      return {
        loadables: map,
        isLoading: loading,
      };
    });
  }, []);

  const value = useMemo(() => {
    return {
      ...state,
      addLoadable,
      removeLoadable,
    };
  }, [state, addLoadable, removeLoadable]);

  return (
    <LoadingContext.Provider value={value}>{children}</LoadingContext.Provider>
  );
}

export function useLoading(isLoading: boolean): void {
  const id = useId();
  // eslint-disable-next-line @typescript-eslint/unbound-method
  const { addLoadable, removeLoadable } = useContext(LoadingContext);
  useEffect(() => {
    addLoadable(id, isLoading);
    return () => {
      removeLoadable(id);
    };
  }, [id, isLoading, addLoadable, removeLoadable]);
}

export function useIsLoading(): boolean {
  const { isLoading } = useContext(LoadingContext);
  return isLoading;
}
