import { createContext, useContext } from "react";

export interface AppContextValue {
  appNodeID: string;
  appID: string;
  currentCollaboratorRole: string;
}

export const AppContext = createContext<AppContextValue | null>(null);

export function useAppContext(): AppContextValue {
  const value = useContext(AppContext);
  if (!value) {
    throw Error("useAppContext must be used within AppContext.Provider");
  }
  return value;
}

export function useOptionalAppContext(): AppContextValue | null {
  return useContext(AppContext);
}
