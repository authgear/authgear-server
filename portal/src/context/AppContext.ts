import { createContext, useContext } from "react";

interface AppContextValue {
  appID: string;
}

export const AppContext = createContext<AppContextValue | null>(null);

export function useAppContext(): AppContextValue {
  const value = useContext(AppContext);
  if (!value) {
    throw Error("useAppContext must be used within AppContext.Provider");
  }
  return value;
}
